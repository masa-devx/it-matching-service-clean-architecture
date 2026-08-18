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

@route("/projects")
interface Projects {
  /** 公開中の案件一覧（seek ページネーション・検索。未公開は含まれない） */
  @useAuth(BearerAuth)
  @get
  list(
    /** 前ページの next_cursor を渡すと、そこから続きを読む */
    @query cursor?: int64,

    /** 取得件数（既定20・最大50にクランプされる） */
    @query limit?: int32,

    /** 指定スキルを全て持つ案件に絞る（AND） */
    @query skills?: string[],

    @query remote_ok?: boolean,

    /** 時給下限がこの額以上の案件に絞る */
    @query min_hourly_rate?: int32,
  ): {
    @statusCode statusCode: 200;
    @body page: ProjectPage;
  } | ApiError;

  /** 公開中の案件詳細（未公開は404） */
  @useAuth(BearerAuth)
  @route("/{id}")
  @get
  get(@path id: int64): {
    @statusCode statusCode: 200;
    @body project: Project;
  } | ApiError;
}

@route("/applications")
interface Applications {
  /** 応募する（公開中の案件のみ・同じ案件への二重応募は409） */
  @useAuth(BearerAuth)
  @post
  create(@body input: ApplicationCreateInput): {
    @statusCode statusCode: 201;
    @body application: Application;
  } | ApiError;

  /** 自分の応募一覧（新しい順・案件タイトル込み） */
  @useAuth(BearerAuth)
  @get
  list(): {
    @statusCode statusCode: 200;
    @body applications: ApplicationList;
  } | ApiError;

  /** 応募を取り下げる（applied / offered から。決着済みは409） */
  @useAuth(BearerAuth)
  @route("/{id}/withdraw")
  @post
  withdraw(@path id: int64): {
    @statusCode statusCode: 200;
    @body application: Application;
  } | ApiError;

  /** オファーを承諾する（offered → accepted。ダブルオプトインの成立） */
  @useAuth(BearerAuth)
  @route("/{id}/accept")
  @post
  accept(@path id: int64): {
    @statusCode statusCode: 200;
    @body application: Application;
  } | ApiError;

  /** オファーを辞退する（offered → declined） */
  @useAuth(BearerAuth)
  @route("/{id}/decline")
  @post
  decline(@path id: int64): {
    @statusCode statusCode: 200;
    @body application: Application;
  } | ApiError;
}
