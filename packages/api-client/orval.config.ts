import { defineConfig } from "orval";

// OpenAPI → TypeScript の生成設定。
// 同じ spec から「型 + Fetch クライアント」と「Zod スキーマ」の2系統を出す
export default defineConfig({
  // 型 + Fetch クライアント（Server Action から Go API を呼ぶのに使う）
  company: {
    input: "../spec/openapi-company.yaml",
    output: {
      target: "company/generated/endpoints.ts",
      schemas: "company/generated/models",
      client: "fetch",
      mode: "split",
      override: {
        // fetch を customFetch に差し替え、ベースURL（＋/company マウント）を注入する
        mutator: {
          path: "custom-fetch.ts",
          name: "customFetch",
        },
      },
    },
  },
  // フォーム検証用の Zod スキーマ（.omit().extend() で加工して使う＝設計プラン§9 案1）
  companyZod: {
    input: "../spec/openapi-company.yaml",
    output: {
      target: "company/generated/zod.ts",
      client: "zod",
    },
  },
  // talent 系統（company と対称・mutator だけ /talent マウント用に切替）
  talent: {
    input: "../spec/openapi-talent.yaml",
    output: {
      target: "talent/generated/endpoints.ts",
      schemas: "talent/generated/models",
      client: "fetch",
      mode: "split",
      override: {
        mutator: {
          path: "custom-fetch.ts",
          name: "customFetchTalent",
        },
      },
    },
  },
  talentZod: {
    input: "../spec/openapi-talent.yaml",
    output: {
      target: "talent/generated/zod.ts",
      client: "zod",
    },
  },
});
