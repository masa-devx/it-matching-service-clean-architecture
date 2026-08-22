// perfseed はパフォーマンス計測用のデモデータを dev DB へ投入する（#114）。
//
// 設計:
//   - id は 1,000,000 番台に明示採番 → 既存の開発データと共存でき、削除も「id >= 100万」だけで安全
//   - 決定的乱数（seed 固定）→ 同じコマンドは常に同じデータ = 計測の再現性
//   - UNNEST バッチ INSERT（1,000行/文）→ 1行ずつの INSERT の数百倍速い
//   - 分布を現実に寄せる（スキルの頻度差・状態の偏り・過去1年の時間分散）
//     一様乱数のデータは本番のインデックス挙動を再現しないため
//
// テスト・基準世界（tsunagu_test / tsunagu_e2e / dump.sql）には一切使わない。
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/config"
)

// idBase 以上がデモデータの帯域（既存の開発シードは 1〜 の低い id を使う）
const idBase = 1_000_000

// e2e fixture と同じ固定 bcrypt ハッシュ（パスワード e2ePassword123・テスト専用の公開値）。
// 毎回 bcrypt を計算すると 1万ユーザーで数分かかるため固定値を使う
const passwordHash = "$2a$10$OW9eXMdehG7ljy2gSTWfleDRMTG57B0QRv6im2FtyPaBp5Ad1UifS" //nolint:gosec

// skillPool は出現頻度つきのスキル一覧（現実の求人票の偏りを模す。一様分布にしない）
var skillPool = []struct {
	name   string
	weight int
}{
	{"Go", 30},
	{"TypeScript", 30},
	{"React", 25},
	{"Next.js", 20},
	{"PostgreSQL", 18},
	{"AWS", 15},
	{"Docker", 12},
	{"Python", 12},
	{"Java", 10},
	{"Kubernetes", 8},
	{"Terraform", 6},
	{"Rust", 4},
}

func main() {
	companies := flag.Int("companies", 1000, "生成する企業数")
	talents := flag.Int("talents", 10000, "生成する人材数")
	projects := flag.Int("projects", 100000, "生成する案件数")
	applications := flag.Int("applications", 300000, "生成する応募数")
	seed := flag.Int64("seed", 1, "乱数シード（同じ値なら同じデータ）")
	yes := flag.Bool("yes", false, "確認プロンプトをスキップ")
	flag.Parse()

	if err := run(*companies, *talents, *projects, *applications, *seed, *yes); err != nil {
		log.Fatalf("perfseed: %v", err)
	}
}

func run(companies, talents, projects, applications int, seed int64, yes bool) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// 事故防止: テスト系 DB への投入を拒否する（デモデータは dev 専用）
	if strings.Contains(cfg.DatabaseURL, "_test") || strings.Contains(cfg.DatabaseURL, "_e2e") {
		return fmt.Errorf("テスト系 DB（_test / _e2e）には投入できません: %s", cfg.DatabaseURL)
	}

	fmt.Printf("投入先: %s\n企業 %d・人材 %d・案件 %d・応募 %d（id は %d 番台・seed=%d）\n",
		cfg.DatabaseURL, companies, talents, projects, applications, idBase, seed)
	if !yes {
		fmt.Print("よろしいですか？ [y/N]: ")
		var answer string
		_, _ = fmt.Scanln(&answer)
		if answer != "y" && answer != "Y" {
			fmt.Println("中止しました")
			return nil
		}
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("DB 接続に失敗: %w", err)
	}
	defer func() { _ = conn.Close(ctx) }()

	// 二重投入ガード: id は明示採番のため、残留データがあると主キー衝突で途中失敗する。
	// 分かりにくい 23505 エラーになる前に検出し、対処コマンドを案内する
	var existing int64
	if err := conn.QueryRow(ctx, `SELECT count(*) FROM users WHERE id >= $1`, idBase).Scan(&existing); err != nil {
		return fmt.Errorf("既存デモデータの確認に失敗: %w", err)
	}
	if existing > 0 {
		return fmt.Errorf("デモデータが既に存在します（users %d 件）。先に make seed-perf-clean を実行してください", existing)
	}

	// 計測の再現性の要: 固定シードの疑似乱数（暗号用途ではないので math/rand でよい）
	rng := rand.New(rand.NewSource(seed)) //nolint:gosec

	start := time.Now()
	if err := seedUsersAndProfiles(ctx, conn, rng, companies, talents); err != nil {
		return err
	}
	publishedIDs, err := seedProjects(ctx, conn, rng, companies, projects)
	if err != nil {
		return err
	}
	if err := seedApplications(ctx, conn, rng, talents, applications, publishedIDs); err != nil {
		return err
	}
	if err := bumpSequences(ctx, conn); err != nil {
		return err
	}

	fmt.Printf("完了: %s\n", time.Since(start).Round(time.Millisecond))
	return nil
}

// baseTime はデータの時間分散の起点（過去1年に散らす。決定的にするため固定値）
var baseTime = time.Date(2025, 8, 21, 0, 0, 0, 0, time.UTC)

// spreadTime は過去1年のどこかの決定的な時刻を返す
func spreadTime(rng *rand.Rand) time.Time {
	return baseTime.Add(time.Duration(rng.Int63n(365*24)) * time.Hour)
}

// pickSkills は頻度重みに従って n 個のスキルを重複なしで選ぶ
func pickSkills(rng *rand.Rand, n int) []string {
	total := 0
	for _, s := range skillPool {
		total += s.weight
	}
	picked := make([]string, 0, n)
	used := make(map[string]bool, n)
	for len(picked) < n {
		r := rng.Intn(total)
		for _, s := range skillPool {
			if r -= s.weight; r < 0 {
				if !used[s.name] {
					used[s.name] = true
					picked = append(picked, s.name)
				}
				break
			}
		}
	}
	return picked
}

// batchSize は 1 INSERT 文に束ねる行数
const batchSize = 1000

func seedUsersAndProfiles(ctx context.Context, conn *pgx.Conn, rng *rand.Rand, companies, talents int) error {
	// users（企業 + 人材。id: 企業=idBase+1..・人材=idBase+companies+1..）
	total := companies + talents
	for off := 0; off < total; off += batchSize {
		n := min(batchSize, total-off)
		ids := make([]int64, n)
		emails := make([]string, n)
		roles := make([]string, n)
		createdAts := make([]time.Time, n)
		for i := range n {
			idx := off + i
			ids[i] = int64(idBase + 1 + idx)
			if idx < companies {
				roles[i] = "company"
				emails[i] = fmt.Sprintf("perf-company-%d@example.com", idx+1)
			} else {
				roles[i] = "talent"
				emails[i] = fmt.Sprintf("perf-talent-%d@example.com", idx-companies+1)
			}
			createdAts[i] = spreadTime(rng)
		}
		_, err := conn.Exec(ctx, `
			INSERT INTO users (id, email, password_hash, role, created_at)
			OVERRIDING SYSTEM VALUE
			SELECT u.id, u.email, $2, u.role, u.created_at
			FROM unnest($1::bigint[], $3::text[], $4::text[], $5::timestamptz[]) AS u(id, email, role, created_at)`,
			ids, passwordHash, emails, roles, createdAts)
		if err != nil {
			return fmt.Errorf("users の投入に失敗（offset %d）: %w", off, err)
		}
	}

	// companies（id = idBase+1..idBase+companies・user_id は同じ並び）
	for off := 0; off < companies; off += batchSize {
		n := min(batchSize, companies-off)
		ids := make([]int64, n)
		userIDs := make([]int64, n)
		names := make([]string, n)
		for i := range n {
			idx := off + i
			ids[i] = int64(idBase + 1 + idx)
			userIDs[i] = int64(idBase + 1 + idx)
			names[i] = fmt.Sprintf("パフォーマンス商事%d", idx+1)
		}
		_, err := conn.Exec(ctx, `
			INSERT INTO companies (id, user_id, name, location, description, created_at)
			OVERRIDING SYSTEM VALUE
			SELECT c.id, c.user_id, c.name, '東京', '計測用のデモ企業です', now()
			FROM unnest($1::bigint[], $2::bigint[], $3::text[]) AS c(id, user_id, name)`,
			ids, userIDs, names)
		if err != nil {
			return fmt.Errorf("companies の投入に失敗（offset %d）: %w", off, err)
		}
	}

	// talents（スキルは頻度分布・1〜5個。配列は '|' 連結で渡して SQL 側で string_to_array）
	for off := 0; off < talents; off += batchSize {
		n := min(batchSize, talents-off)
		ids := make([]int64, n)
		userIDs := make([]int64, n)
		names := make([]string, n)
		skills := make([]string, n)
		for i := range n {
			idx := off + i
			ids[i] = int64(idBase + 1 + idx)
			userIDs[i] = int64(idBase + 1 + companies + idx)
			names[i] = fmt.Sprintf("計測人材%d", idx+1)
			skills[i] = strings.Join(pickSkills(rng, 1+rng.Intn(5)), "|")
		}
		_, err := conn.Exec(ctx, `
			INSERT INTO talents (id, user_id, display_name, skills, bio, created_at)
			OVERRIDING SYSTEM VALUE
			SELECT t.id, t.user_id, t.display_name, string_to_array(t.skills, '|'), '計測用のデモ人材です', now()
			FROM unnest($1::bigint[], $2::bigint[], $3::text[], $4::text[]) AS t(id, user_id, display_name, skills)`,
			ids, userIDs, names, skills)
		if err != nil {
			return fmt.Errorf("talents の投入に失敗（offset %d）: %w", off, err)
		}
	}

	fmt.Printf("users %d・companies %d・talents %d 投入\n", total, companies, talents)
	return nil
}

// seedProjects は案件を投入し、応募の対象になる公開案件の id 一覧を返す
func seedProjects(ctx context.Context, conn *pgx.Conn, rng *rand.Rand, companies, projects int) ([]int64, error) {
	var publishedIDs []int64
	for off := 0; off < projects; off += batchSize {
		n := min(batchSize, projects-off)
		ids := make([]int64, n)
		companyIDs := make([]int64, n)
		titles := make([]string, n)
		statuses := make([]string, n)
		hours := make([]int32, n)
		remotes := make([]bool, n)
		skills := make([]string, n)
		rateMins := make([]int32, n) // -1 = NULL（未設定4割: NULL の分布も現実に寄せる）
		rateMaxs := make([]int32, n)
		createdAts := make([]time.Time, n)
		for i := range n {
			idx := off + i
			ids[i] = int64(idBase + 1 + idx)
			companyIDs[i] = int64(idBase + 1 + rng.Intn(companies))
			titles[i] = fmt.Sprintf("計測案件%d: %s の開発支援", idx+1, pickSkills(rng, 1)[0])
			// 公開7割・下書き2割・終了1割（一覧・検索の対象は公開のみ = 現実の比率が計測に効く）
			switch r := rng.Intn(10); {
			case r < 7:
				statuses[i] = "published"
				publishedIDs = append(publishedIDs, ids[i])
			case r < 9:
				statuses[i] = "draft"
			default:
				statuses[i] = "closed"
			}
			hours[i] = int32(5 + rng.Intn(6)*5) //nolint:gosec // 値域 5〜30
			remotes[i] = rng.Intn(10) < 6       // リモート可6割
			skills[i] = strings.Join(pickSkills(rng, 1+rng.Intn(4)), "|")
			if rng.Intn(10) < 6 {
				lo := int32(3000 + rng.Intn(50)*100) //nolint:gosec // 値域 3000〜7900
				rateMins[i] = lo
				rateMaxs[i] = lo + int32(500+rng.Intn(30)*100) //nolint:gosec // 値域 500〜3400
			} else {
				rateMins[i], rateMaxs[i] = -1, -1
			}
			createdAts[i] = spreadTime(rng)
		}
		_, err := conn.Exec(ctx, `
			INSERT INTO projects (id, company_id, title, description, status, hours_per_week, remote_ok, required_skills, hourly_rate_min, hourly_rate_max, created_at)
			OVERRIDING SYSTEM VALUE
			SELECT p.id, p.company_id, p.title, '計測用のデモ案件です', p.status, p.hours, p.remote,
			       string_to_array(p.skills, '|'), NULLIF(p.rate_min, -1), NULLIF(p.rate_max, -1), p.created_at
			FROM unnest($1::bigint[], $2::bigint[], $3::text[], $4::text[], $5::int[], $6::bool[], $7::text[], $8::int[], $9::int[], $10::timestamptz[])
			     AS p(id, company_id, title, status, hours, remote, skills, rate_min, rate_max, created_at)`,
			ids, companyIDs, titles, statuses, hours, remotes, skills, rateMins, rateMaxs, createdAts)
		if err != nil {
			return nil, fmt.Errorf("projects の投入に失敗（offset %d）: %w", off, err)
		}
	}
	fmt.Printf("projects %d 投入（うち公開 %d）\n", projects, len(publishedIDs))
	return publishedIDs, nil
}

// bumpSequences は明示採番した id をシーケンスが追い越すように前方調整する
// （以後のアプリからの INSERT がデモデータの id と衝突しないため）
func bumpSequences(ctx context.Context, conn *pgx.Conn) error {
	for _, table := range []string{"users", "companies", "talents", "projects", "applications"} {
		q := fmt.Sprintf(`SELECT setval(
			pg_get_serial_sequence('%[1]s', 'id'),
			GREATEST((SELECT COALESCE(max(id), 1) FROM %[1]s),
			         COALESCE(pg_sequence_last_value(pg_get_serial_sequence('%[1]s', 'id')), 1)))`, table)
		if _, err := conn.Exec(ctx, q); err != nil {
			return fmt.Errorf("%s のシーケンス調整に失敗: %w", table, err)
		}
	}
	return nil
}

// applicationStatuses は応募の状態分布（applied 50%・offered 15%・accepted 10%・
// rejected 15%・withdrawn 5%・declined 5%。遷移表で到達可能な状態のみを使う）。
// 基準世界と違い usecase を通さず直接 INSERT する: 目的が読み取り計測であり、
// 30万件を遷移表経由で作るのは実用にならないため（線引きは docs/DB.md 参照）
var applicationStatuses = []struct {
	name         string
	weight       int
	companyActed bool // company_acted_at を持つ状態か
	talentActed  bool // talent_acted_at を持つ状態か
}{
	{"applied", 50, false, false},
	{"offered", 15, true, false},
	{"accepted", 10, true, true},
	{"rejected", 15, true, false},
	{"withdrawn", 5, false, true},
	{"declined", 5, true, true},
}

// seedApplications は応募を投入する。
// (talent_id, project_id) は UNIQUE のため、人材ごとに「開始位置 + 等間隔ストライド」で
// 公開案件を選ぶ（step*件数 < 公開数 なら重複が起きない決定的な組み合わせ生成）
func seedApplications(ctx context.Context, conn *pgx.Conn, rng *rand.Rand, talents, applications int, publishedIDs []int64) error {
	if len(publishedIDs) == 0 {
		return fmt.Errorf("公開案件が無いため応募を生成できません")
	}
	perTalent := applications / talents
	if perTalent >= len(publishedIDs) {
		return fmt.Errorf("人材あたり応募数 %d が公開案件数 %d を超えています", perTalent, len(publishedIDs))
	}
	// 等間隔ストライド: j*step (j < perTalent) が公開数を超えない = 1人の中で必ず別案件
	step := (len(publishedIDs) - 1) / perTalent

	total := talents * perTalent
	statusTotal := 0
	for _, st := range applicationStatuses {
		statusTotal += st.weight
	}

	for off := 0; off < total; off += batchSize {
		n := min(batchSize, total-off)
		ids := make([]int64, n)
		talentIDs := make([]int64, n)
		projectIDs := make([]int64, n)
		statuses := make([]string, n)
		messages := make([]string, n)
		companyActs := make([]string, n) // '' = NULL（timestamptz の NULL は空文字センチネル + NULLIF）
		talentActs := make([]string, n)
		createdAts := make([]time.Time, n)
		for i := range n {
			idx := off + i
			t := idx / perTalent // 何番目の人材か
			j := idx % perTalent // その人材の何件目の応募か
			ids[i] = int64(idBase + 1 + idx)
			talentIDs[i] = int64(idBase + 1 + t)
			projectIDs[i] = publishedIDs[(t+j*step)%len(publishedIDs)]

			r := rng.Intn(statusTotal)
			for _, st := range applicationStatuses {
				if r -= st.weight; r < 0 {
					statuses[i] = st.name
					created := spreadTime(rng)
					createdAts[i] = created
					if st.companyActed {
						companyActs[i] = created.Add(30 * time.Minute).Format(time.RFC3339)
					}
					if st.talentActed {
						talentActs[i] = created.Add(60 * time.Minute).Format(time.RFC3339)
					}
					break
				}
			}
			messages[i] = fmt.Sprintf("計測用の応募です（%d件目）", idx+1)
		}
		_, err := conn.Exec(ctx, `
			INSERT INTO applications (id, talent_id, project_id, status, message, company_acted_at, talent_acted_at, created_at)
			OVERRIDING SYSTEM VALUE
			SELECT a.id, a.talent_id, a.project_id, a.status, a.message,
			       NULLIF(a.company_acted, '')::timestamptz, NULLIF(a.talent_acted, '')::timestamptz, a.created_at
			FROM unnest($1::bigint[], $2::bigint[], $3::bigint[], $4::text[], $5::text[], $6::text[], $7::text[], $8::timestamptz[])
			     AS a(id, talent_id, project_id, status, message, company_acted, talent_acted, created_at)`,
			ids, talentIDs, projectIDs, statuses, messages, companyActs, talentActs, createdAts)
		if err != nil {
			return fmt.Errorf("applications の投入に失敗（offset %d）: %w", off, err)
		}
	}
	fmt.Printf("applications %d 投入（人材あたり %d 件・ストライド %d）\n", total, perTalent, step)
	return nil
}
