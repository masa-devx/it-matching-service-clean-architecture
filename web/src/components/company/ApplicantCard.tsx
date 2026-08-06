import { Briefcase } from 'lucide-react'
import type { CompanyApplication } from '@/lib/companyApplications'
import {
  ApplicationStatusBadge,
  applicationStatusDescription,
} from '@/components/ApplicationStatusBadge'
import { ApplicantActions } from '@/components/company/ApplicantActions'
import { Badge } from '@/components/ui/badge'

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('ja-JP', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })
}

export function ApplicantCard({
  application,
}: {
  application: CompanyApplication
}) {
  return (
    <div className="flex flex-col gap-4 rounded-lg border bg-card p-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="flex flex-col gap-1">
          {/* 連絡先は表示しない（契約成立まで直接やり取りできない設計） */}
          <h2 className="font-bold">{application.display_name}</h2>
          <p className="text-sm text-muted-foreground">
            {formatDate(application.created_at)}に応募
          </p>
        </div>
        <ApplicationStatusBadge status={application.status} />
      </div>

      {/* 選考の判断材料（経験年数・スキル）を志望動機より先に置く */}
      <div className="flex flex-wrap items-center gap-x-4 gap-y-2 text-sm">
        <span className="flex items-center gap-1.5 text-muted-foreground">
          <Briefcase className="size-4" aria-hidden="true" />
          実務{application.years_of_exp}年
        </span>
        {application.skills.length > 0 && (
          <ul className="flex flex-wrap gap-1.5">
            {application.skills.map((skill) => (
              <li key={skill}>
                <Badge variant="secondary">{skill}</Badge>
              </li>
            ))}
          </ul>
        )}
      </div>

      <p className="text-sm text-muted-foreground">
        {applicationStatusDescription(application.status)}
      </p>

      {/* 志望動機は長くなるため折りたたむ。選考で必ず読むので既定で開いておく */}
      <details open className="text-sm">
        <summary className="w-fit cursor-pointer text-muted-foreground hover:text-foreground">
          志望動機
        </summary>
        <p className="mt-2 whitespace-pre-wrap break-words leading-relaxed">
          {application.message}
        </p>
      </details>

      <ApplicantActions
        applicationId={application.id}
        status={application.status}
      />
    </div>
  )
}
