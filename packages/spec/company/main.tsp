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

@route("/projects")
interface Projects {
  /** 案件を作成する（draft として作成される） */
  @post
  create(@body input: ProjectCreateInput): {
    @statusCode statusCode: 201;
    @body project: Project;
  } | ApiError;
}
