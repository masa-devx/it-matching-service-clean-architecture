// orval の mutator（生成コードが fetch の代わりに呼ぶ関数）。
// 役割はベースURLの注入だけに絞る（認証ヘッダー等は Phase 2 でここに集約する）。
//
// 仕様のパスは「company API の中身」（/projects）だけを表し、/company プレフィックスは
// サーバー側のマウント（main.go の BaseURL）と対で、クライアント側ではここが対応する。
const baseUrl = () => {
  const origin = process.env.API_URL ?? "http://localhost:8082";
  return `${origin}/company`;
};

export const customFetch = async <T>(
  url: string,
  options: RequestInit,
): Promise<T> => {
  const res = await fetch(`${baseUrl()}${url}`, options);

  const body = [204, 205, 304].includes(res.status) ? null : await res.text();
  const data = body ? JSON.parse(body) : {};

  return { data, status: res.status, headers: res.headers } as T;
};
