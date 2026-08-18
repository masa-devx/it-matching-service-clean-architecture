package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	company "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/api/company"
	talentapi "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/api/talent"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/db"
	companyhandler "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/company/handler"
	companyusecase "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/company/usecase"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/auth"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/config"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/infra"
	talenthandler "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/talent/handler"
	talentusecase "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/talent/usecase"
)

// main は組み立て（DI）だけを行う: 設定読込 → DB接続 → 依存の手渡し → ルーター登録 → 起動。
// ルーティングは生成コードに任せ、手書きのルートを増やさない（health は運用エンドポイントなので例外）
func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx := context.Background()
	pool, err := infra.NewDB(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	queries := db.New(pool)
	projectUsecase := companyusecase.NewProject(queries)
	authUsecase := companyusecase.NewAuth(pool, queries)
	companyApplicationUsecase := companyusecase.NewApplication(queries)

	// 認証必須の operation は仕様（security 定義）から起動時に1回だけ導出する。
	// コードに手書きのリストを持たない＝認証要否の一次情報は .tsp の @useAuth
	companySpec, err := company.GetSpec()
	if err != nil {
		return fmt.Errorf("company 仕様の読み込みに失敗: %w", err)
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
		companyhandler.New(projectUsecase, authUsecase, companyApplicationUsecase, cfg.JWTSecret),
		[]company.StrictMiddlewareFunc{auth.NewStrictAuth[company.StrictHandlerFunc](cfg.JWTSecret, auth.RoleCompany, companyAuthOps)},
	)
	company.HandlerWithOptions(companyStrict, company.StdHTTPServerOptions{
		BaseURL:    "/company",
		BaseRouter: mux,
	})

	// talent API: company と対称のマウント（/talent × role=talent を一律強制）
	talentAuthUsecase := talentusecase.NewAuth(pool, queries)
	talentProjectUsecase := talentusecase.NewProject(queries)
	talentApplicationUsecase := talentusecase.NewApplication(queries)
	talentSpec, err := talentapi.GetSpec()
	if err != nil {
		return fmt.Errorf("talent 仕様の読み込みに失敗: %w", err)
	}
	talentAuthOps := auth.RequiredAuthOps(talentSpec)
	talentStrict := talentapi.NewStrictHandler(
		talenthandler.New(talentAuthUsecase, talentProjectUsecase, talentApplicationUsecase, cfg.JWTSecret),
		[]talentapi.StrictMiddlewareFunc{auth.NewStrictAuth[talentapi.StrictHandlerFunc](cfg.JWTSecret, auth.RoleTalent, talentAuthOps)},
	)
	talentapi.HandlerWithOptions(talentStrict, talentapi.StdHTTPServerOptions{
		BaseURL:    "/talent",
		BaseRouter: mux,
	})

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second, // Slowloris 対策（ヘッダーを送り切らない接続を切る）
	}

	log.Printf("api-server listening on :%d", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
