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
  /** 案件を作成する（draft として作成される・要認証。所有者はトークンから決まる） */
  @useAuth(BearerAuth)
  @post
  create(@body input: ProjectCreateInput): {
    @statusCode statusCode: 201;
    @body project: Project;
  } | ApiError;

  /** 公開する（draft / closed → published。遷移の可否は掲載状態の遷移表が決める） */
  @useAuth(BearerAuth)
  @route("/{id}/publish")
  @post
  publish(@path id: int64): {
    @statusCode statusCode: 200;
    @body project: Project;
  } | ApiError;

  /** 非公開にする（published → draft） */
  @useAuth(BearerAuth)
  @route("/{id}/unpublish")
  @post
  unpublish(@path id: int64): {
    @statusCode statusCode: 200;
    @body project: Project;
  } | ApiError;

  /** 募集を終了する（published → closed。再募集は publish） */
  @useAuth(BearerAuth)
  @route("/{id}/close")
  @post
  close(@path id: int64): {
    @statusCode statusCode: 200;
    @body project: Project;
  } | ApiError;
}
