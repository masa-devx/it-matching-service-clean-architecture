import Link from 'next/link'
import { FileText } from 'lucide-react'
import {
  getMyContracts,
  currentContractPage,
  CONTRACTS_PER_PAGE,
  type ContractSearchParams,
} from '@/lib/contracts'
import {
  ContractStatusBadge,
  contractStatusDescription,
} from '@/components/ContractStatusBadge'
import { ContractStatusFilter } from '@/components/ContractStatusFilter'
import { PageHeader } from '@/components/PageHeader'
import { EmptyState } from '@/components/EmptyState'
import { Pagination } from '@/components/Pagination'
import { Button } from '@/components/ui/button'

export const metadata = { title: '契約 | Tsunagu Works' }

export default async function CompanyContractsPage({
  searchParams,
}: {
  searchParams: Promise<ContractSearchParams>
}) {
  const params = await searchParams
  const result = await getMyContracts(params)

  if (!result) {
    return <p className="text-destructive">契約を取得できませんでした</p>
  }

  const page = currentContractPage(params)
  const filtered = Boolean(params.status)

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="契約"
        description={`${result.total.toLocaleString()}件の契約`}
      />

      <ContractStatusFilter
        basePath="/company/contracts"
        current={params.status}
      />

      {result.contracts.length === 0 ? (
        <EmptyState
          icon={FileText}
          title={filtered ? 'この状態の契約はありません' : '契約がありません'}
          description={
            filtered
              ? '別の状態を選ぶか、すべて表示してください'
              : '応募を承諾すると、ここに契約が表示されます'
          }
          action={
            <Button asChild className="h-11">
              <Link
                href={filtered ? '/company/contracts' : '/company/projects'}
              >
                {filtered ? 'すべて表示' : '案件管理へ'}
              </Link>
            </Button>
          }
        />
      ) : (
        <ul className="flex flex-col gap-3">
          {result.contracts.map((contract) => (
            <li key={contract.id}>
              <Link
                href={`/company/contracts/${contract.id}`}
                className="flex flex-col gap-3 rounded-lg border bg-card p-4 transition-colors hover:border-primary hover:bg-primary/5"
              >
                <div className="flex flex-wrap items-center gap-2">
                  <ContractStatusBadge status={contract.status} />
                  <h2 className="font-bold leading-snug">{contract.title}</h2>
                </div>

                {/* 企業から見た相手は人材。条件は承諾時点のスナップショット */}
                <p className="text-sm text-muted-foreground">
                  {contract.talent_name}・
                  <span className="tabular-nums">
                    {contract.hourly_rate.toLocaleString()}円/時
                  </span>
                  ・週{contract.hours_per_week}時間
                </p>

                <p className="text-sm text-muted-foreground">
                  {contractStatusDescription(contract.status)}
                </p>

                {/* 企業が一覧で最も知りたいのは「確認していない報告があるか」 */}
                <p className="flex flex-wrap items-center gap-1.5 text-sm">
                  <span className="tabular-nums text-muted-foreground">
                    稼働報告 {contract.work_report_count}件
                  </span>
                  {contract.pending_report_count > 0 && (
                    <span className="font-medium tabular-nums text-primary">
                      （{contract.pending_report_count}件が未確認）
                    </span>
                  )}
                </p>
              </Link>
            </li>
          ))}
        </ul>
      )}

      <Pagination
        currentPage={page}
        total={result.total}
        perPage={CONTRACTS_PER_PAGE}
      />
    </div>
  )
}
