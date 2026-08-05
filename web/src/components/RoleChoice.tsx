import Link from 'next/link'
import { Building2, User } from 'lucide-react'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'

// 企業/人材の入口を選ぶ画面。ログインと新規登録で遷移先だけが変わるため、
// action（'login' | 'signup'）で出し分ける
export function RoleChoice({ action }: { action: 'login' | 'signup' }) {
  const isSignup = action === 'signup'
  const choices = [
    {
      href: `/company/${action}`,
      icon: Building2,
      title: '企業の方',
      description: isSignup
        ? '案件を掲載して人材を探す'
        : '掲載した案件を管理する',
    },
    {
      href: `/talent/${action}`,
      icon: User,
      title: '人材の方',
      description: isSignup ? '案件を探して応募する' : '応募状況を確認する',
    },
  ]

  return (
    <Card className="w-full max-w-md">
      <CardHeader>
        <CardTitle>{isSignup ? '新規登録' : 'ログイン'}</CardTitle>
        <CardDescription>ご利用の立場を選んでください</CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col gap-3">
        {choices.map(({ href, icon: Icon, title, description }) => (
          <Link
            key={href}
            href={href}
            className="flex items-center gap-4 rounded-lg border p-4 transition-colors hover:border-primary hover:bg-primary/5"
          >
            <Icon className="size-6 text-primary" aria-hidden="true" />
            <div className="flex flex-col gap-0.5">
              <span className="font-medium">{title}</span>
              <span className="text-sm text-muted-foreground">
                {description}
              </span>
            </div>
          </Link>
        ))}
      </CardContent>
    </Card>
  )
}
