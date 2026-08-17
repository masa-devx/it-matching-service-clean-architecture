// orval の mutator（生成コードが fetch の代わりに呼ぶ関数）。
// 役割はベースURLの注入だけに絞る（認証ヘッダー等は Phase 2 でここに集約する）。
//
// 仕様のパスは各 API の「中身」（/auth/... 等）だけを表し、/company・/talent の
// プレフィックスはサーバー側のマウント（main.go の BaseURL）と対で、ここが対応する
const createFetch =
  (mountPath: string) =>
  async <T>(url: string, options: RequestInit): Promise<T> => {
    const origin = process.env.API_URL ?? "http://localhost:8082";
    const res = await fetch(`${origin}${mountPath}${url}`, options);

    const body = [204, 205, 304].includes(res.status)
      ? null
      : await res.text();
    const data = body ? JSON.parse(body) : {};

    return { data, status: res.status, headers: res.headers } as T;
  };

export const customFetch = createFetch("/company");
export const customFetchTalent = createFetch("/talent");
