import "@typespec/http";

using TypeSpec.Http;

namespace TsunaguWorks;

/** 案件の掲載状態 */
union ProjectStatus {
  draft: "draft",
  published: "published",
  closed: "closed",
}

/** 案件（企業が掲載する仕事） */
model Project {
  id: int64;

  @maxLength(100)
  title: string;

  @maxLength(5000)
  description: string;

  /** 想定時給の下限（円）。未設定は null */
  @minValue(0)
  hourly_rate_min: int32 | null;

  /** 想定時給の上限（円）。未設定は null */
  @minValue(0)
  hourly_rate_max: int32 | null;

  /** 週の想定稼働時間 */
  @minValue(1)
  @maxValue(60)
  hours_per_week: int32;

  remote_ok: boolean;

  /** 必須スキル（例: ["Go", "PostgreSQL"]） */
  required_skills: string[];

  status: ProjectStatus;

  created_at: utcDateTime;
}

/** 案件の作成入力（status はサーバー側で draft 固定のため含めない） */
model ProjectCreateInput {
  @minLength(1)
  @maxLength(100)
  title: string;

  @minLength(1)
  @maxLength(5000)
  description: string;

  @minValue(0)
  hourly_rate_min?: int32;

  @minValue(0)
  hourly_rate_max?: int32;

  @minValue(1)
  @maxValue(60)
  hours_per_week: int32;

  remote_ok: boolean;

  required_skills: string[];
}

/** 案件の編集入力（作成と同じ形を spread で再利用。status を含めない＝編集で公開状態は変わらない） */
model ProjectUpdateInput {
  ...ProjectCreateInput;
}

/** 案件一覧のレスポンス */
model ProjectList {
  projects: Project[];
}

/** 公開案件の1ページ（seek ページネーション） */
model ProjectPage {
  projects: Project[];

  /** 次ページの読み出し起点となる id。null なら最終ページ */
  next_cursor: int64 | null;
}

/** エラーレスポンスの共通形 */
@error
model ApiError {
  error: string;
  request_id?: string;
}

/** 認証トークン（web は httpOnly Cookie に変換して保持し、ブラウザ JS には渡さない） */
model AuthToken {
  token: string;
}

/** 企業のサインアップ入力（アカウントとプロフィールを同時に作る） */
model CompanySignupInput {
  @format("email")
  email: string;

  /** パスワード（上限は bcrypt の入力上限 72 バイトに合わせる） */
  @secret
  @minLength(8)
  @maxLength(72)
  password: string;

  /** 会社名 */
  @minLength(1)
  @maxLength(100)
  name: string;

  @maxLength(100)
  location?: string;

  @maxLength(1000)
  description?: string;
}

/** ログイン入力 */
model LoginInput {
  @format("email")
  email: string;

  @secret
  password: string;
}

/** ログイン中の企業ユーザー情報 */
model CompanyMe {
  user_id: int64;
  email: string;
  name: string;
  location: string;
  description: string;
}

/** 人材のサインアップ入力（アカウントとプロフィールを同時に作る） */
model TalentSignupInput {
  @format("email")
  email: string;

  /** パスワード（上限は bcrypt の入力上限 72 バイトに合わせる） */
  @secret
  @minLength(8)
  @maxLength(72)
  password: string;

  /** 表示名 */
  @minLength(1)
  @maxLength(50)
  display_name: string;

  /** スキル（例: ["Go", "PostgreSQL"]） */
  skills: string[];

  @maxLength(1000)
  bio?: string;
}

/** ログイン中の人材ユーザー情報 */
model TalentMe {
  user_id: int64;
  email: string;
  display_name: string;
  skills: string[];
  bio: string;
}

/** 応募の選考状態（遷移の許可は shared/domain の遷移表が一次情報） */
union ApplicationStatus {
  applied: "applied",
  offered: "offered",
  accepted: "accepted",
  rejected: "rejected",
  withdrawn: "withdrawn",
  declined: "declined",
}

/** 応募（talent 視点。project_title は projects との JOIN で供給する） */
model Application {
  id: int64;
  project_id: int64;

  /** 応募先案件のタイトル（一覧表示用） */
  project_title: string;

  status: ApplicationStatus;

  /** 志望動機 */
  message: string;

  created_at: utcDateTime;
}

/** 応募の作成入力（talent_id はトークンから・status は applied 固定のため含めない） */
model ApplicationCreateInput {
  project_id: int64;

  /** 志望動機（任意） */
  @maxLength(2000)
  message?: string;
}

/** 応募一覧のレスポンス */
model ApplicationList {
  applications: Application[];
}
