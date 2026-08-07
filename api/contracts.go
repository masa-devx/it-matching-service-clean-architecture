package main

import (
	"database/sql"
	"fmt"
)

// createContractFromApplication は承諾された応募から契約を作る。
//
// 呼び出し側と同じトランザクション（tx）で実行すること。応募の承諾と契約の作成は
// 「合意が成立した」という1つの事実の裏表であり、片方だけ成功する状態を作ってはいけない
// （人材は承諾したつもりなのに契約が無い、という手で直すしかない状態になる）。
//
// 案件の条件は参照ではなく値でコピーする。案件は掲載後に編集できる（PUT /projects/{id}）ため、
// 参照のままだと契約成立後に単価を書き換えられてしまう
func createContractFromApplication(tx *sql.Tx, applicationID int64) error {
	_, err := tx.Exec(
		`INSERT INTO contracts
		   (application_id, project_id, company_id, talent_id,
		    title, hourly_rate, hours_per_week, remote_ok)
		 SELECT a.id, p.id, p.company_id, a.talent_id,
		        p.title, p.hourly_rate_max, p.hours_per_week, p.remote_ok
		 FROM applications a
		 JOIN projects p ON p.id = a.project_id
		 WHERE a.id = $1`,
		applicationID,
	)
	if err != nil {
		return fmt.Errorf("契約の作成に失敗: %w", err)
	}
	return nil
}
