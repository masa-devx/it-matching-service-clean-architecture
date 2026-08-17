import "@typespec/http";

using TypeSpec.Http;

namespace TsunaguWorks;

/** 案件の掲載状態 */
union ProjectStatus {
  draft: "draft",
  published: "published",
  closed: "closed",
}

/** 案件（企業が掲載する仕事）※生成し忘れ実演用の変更・revert予定 */
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

/** エラーレスポンスの共通形 */
@error
model ApiError {
  error: string;
  request_id?: string;
}
