// Package app は組み立て（composition root）を置く。
// 本番（cmd/server）とテスト（test/apitest）が同じ配線を共有するための抽出（#93）。
// ここに無い配線はテストされない＝配線は必ずこの関数に書く
package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	company "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/api/company"
	talentapi "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/api/talent"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/db"
	companyhandler "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/company/handler"
	companyusecase "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/company/usecase"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/auth"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/middleware"
	talenthandler "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/talent/handler"
	talentusecase "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/talent/usecase"
)

// TxBeginner はトランザクションを開始できるもの。
// 本番は *pgxpool.Pool、テストは pgx.Tx（SAVEPOINT の入れ子・ADR-0008）を渡す
type TxBeginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// NewRouter は API サーバーの全配線を組み立てる:
// 生成ルーター2系統（/company・/talent）のマウント・仕様由来の認証ミドルウェア・
// 運用エンドポイント（/health）・共通ミドルウェアチェーン。
// ルーティングは生成コードに任せ、手書きのルートを増やさない（health は運用なので例外）
func NewRouter(txdb TxBeginner, queries *db.Queries, jwtSecret []byte) (http.Handler, error) {
	projectUsecase := companyusecase.NewProject(queries)
	authUsecase := companyusecase.NewAuth(txdb, queries)
	companyApplicationUsecase := companyusecase.NewApplication(queries)

	// 認証必須の operation は仕様（security 定義）から起動時に1回だけ導出する。
	// コードに手書きのリストを持たない＝認証要否の一次情報は .tsp の @useAuth
	companySpec, err := company.GetSpec()
	if err != nil {
		return nil, fmt.Errorf("company 仕様の読み込みに失敗: %w", err)
	}
	companyAuthOps := auth.RequiredAuthOps(companySpec)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// company API: 仕様（openapi-company.yaml）から生成されたルーターを /company 配下にマウントする。
	// パスの一次情報は仕様側にあり、ここでは「どこに載せるか」だけを決める
	companyStrict := company.NewStrictHandler(
		companyhandler.New(projectUsecase, authUsecase, companyApplicationUsecase, jwtSecret),
		[]company.StrictMiddlewareFunc{auth.NewStrictAuth[company.StrictHandlerFunc](jwtSecret, auth.RoleCompany, companyAuthOps)},
	)
	company.HandlerWithOptions(companyStrict, company.StdHTTPServerOptions{
		BaseURL:    "/company",
		BaseRouter: mux,
	})

	// talent API: company と対称のマウント（/talent × role=talent を一律強制）
	talentAuthUsecase := talentusecase.NewAuth(txdb, queries)
	talentProjectUsecase := talentusecase.NewProject(queries)
	talentApplicationUsecase := talentusecase.NewApplication(queries)
	talentSpec, err := talentapi.GetSpec()
	if err != nil {
		return nil, fmt.Errorf("talent 仕様の読み込みに失敗: %w", err)
	}
	talentAuthOps := auth.RequiredAuthOps(talentSpec)
	talentStrict := talentapi.NewStrictHandler(
		talenthandler.New(talentAuthUsecase, talentProjectUsecase, talentApplicationUsecase, jwtSecret),
		[]talentapi.StrictMiddlewareFunc{auth.NewStrictAuth[talentapi.StrictHandlerFunc](jwtSecret, auth.RoleTalent, talentAuthOps)},
	)
	talentapi.HandlerWithOptions(talentStrict, talentapi.StdHTTPServerOptions{
		BaseURL:    "/talent",
		BaseRouter: mux,
	})

	// ミドルウェアチェーン（外→内）: otelhttp → RequestID → Recovery → AccessLog → ルーター。
	// otelhttp が最外なのは、traceparent の抽出と span の開始を最初に行い、
	// 内側の全処理（request_id 付与・ログ・panic 記録）が span の中に入るようにするため。
	// TracerProvider 未設定（テスト・OTLP 無効時）では no-op として振る舞う
	handler := middleware.RequestID(middleware.Recovery(middleware.AccessLog(mux)))
	return otelhttp.NewHandler(handler, "http.server",
		// span 名は「METHOD /path」（既定の固定名だと Jaeger で全リクエストが同名になり探せない）
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
			return r.Method + " " + r.URL.Path
		}),
	), nil
}
