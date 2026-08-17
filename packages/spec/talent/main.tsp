import "@typespec/http";
import "../shared/models.tsp";

using TypeSpec.Http;
using TsunaguWorks;

/**
 * 人材向け API。/talent プレフィックスはサーバー側のマウント時に付与する
 * （company と対称の構成。共通モデルは shared/models.tsp の import で再利用する）
 */
@service(#{ title: "Tsunagu Works Talent API" })
namespace TsunaguWorks.Talent;

@route("/auth")
interface Auth {
  /** サインアップ（アカウントとプロフィールを同時作成し、トークンを発行する） */
  @route("/signup")
  @post
  signup(@body input: TalentSignupInput): {
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
    @body me: TalentMe;
  } | ApiError;
}
