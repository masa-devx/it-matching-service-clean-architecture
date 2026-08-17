import "@typespec/http";
import "../shared/models.tsp";

using TypeSpec.Http;
using TsunaguWorks;

/**
 * 企業向け API。/company プレフィックスはサーバー側のマウント時に付与する
 * （仕様は「company API の中身」だけを定義し、どこに載せるかは実装の関心事とする）
 */
@service(#{ title: "Tsunagu Works Company API" })
namespace TsunaguWorks.Company;

@route("/auth")
interface Auth {
  /** サインアップ（アカウントとプロフィールを同時作成し、トークンを発行する） */
  @route("/signup")
  @post
  signup(@body input: CompanySignupInput): {
    @statusCode statusCode: 201;
    @body auth: AuthToken;
  } | ApiError;

  /** ログイン */
  @route("/login")
  @post
  login(@body input: LoginInput): {
    @statusCode statusCode: 200;
    @body auth: AuthToken;
  } | ApiError;

  /** ログイン中ユーザーの情報（要認証） */
  @useAuth(BearerAuth)
  @route("/me")
  @get
  me(): {
    @statusCode statusCode: 200;
    @body me: CompanyMe;
  } | ApiError;
}

@route("/projects")
interface Projects {
  /** 案件を作成する（draft として作成される） */
  @post
  create(@body input: ProjectCreateInput): {
    @statusCode statusCode: 201;
    @body project: Project;
  } | ApiError;
}
