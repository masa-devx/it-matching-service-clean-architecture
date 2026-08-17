package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	company "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/api/company"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/db"
	companyhandler "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/company/handler"
	companyusecase "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/company/usecase"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/auth"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/config"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/infra"
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

	// 認証が不要な operation のホワイトリスト。
	// TODO(#31): 仕様（OpenAPI の security 定義）から起動時に導出する形へ置き換える
	publicOps := map[string]bool{
		"AuthSignup":     true,
		"AuthLogin":      true,
		"ProjectsCreate": true, // #31 で保護に切り替え（company_id をトークンから取るのと同時）
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// company API: 仕様（openapi-company.yaml）から生成されたルーターを /company 配下にマウントする。
	// パスの一次情報は仕様側にあり、ここでは「どこに載せるか」だけを決める
	companyStrict := company.NewStrictHandler(
		companyhandler.New(projectUsecase, authUsecase, cfg.JWTSecret),
		[]company.StrictMiddlewareFunc{auth.NewStrictAuth[company.StrictHandlerFunc](cfg.JWTSecret, publicOps)},
	)
	company.HandlerWithOptions(companyStrict, company.StdHTTPServerOptions{
		BaseURL:    "/company",
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
