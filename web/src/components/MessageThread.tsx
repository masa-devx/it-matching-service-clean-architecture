import { MessagesSquare, ShieldAlert } from 'lucide-react'
import type { Message } from '@/lib/messages'
import type { CurrentUser } from '@/lib/auth'
import { EmptyState } from '@/components/EmptyState'

function formatDateTime(iso: string): string {
  return new Date(iso).toLocaleString('ja-JP', {
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

// 契約の当事者同士の会話。企業・人材で同じものを表示する
// （共有された記録なので視点で見え方を変えない。契約・稼働報告と同じ判断）。
//
// 自分の発言かどうかは currentRole と sender_role の比較で決まる。
// 配置と背景色で区別するが、**色だけに頼らず送信者名も必ず表示する**
export function MessageThread({
  messages,
  currentRole,
}: {
  messages: Message[]
  currentRole: CurrentUser['role']
}) {
  if (messages.length === 0) {
    return (
      <EmptyState
        icon={MessagesSquare}
        title="メッセージはまだありません"
        description="作業の相談や進捗の共有に使えます"
      />
    )
  }

  return (
    <ul className="flex flex-col gap-4">
      {messages.map((message) => {
        const isMine = message.sender_role === currentRole

        return (
          <li
            key={message.id}
            className={`flex flex-col gap-1 ${isMine ? 'items-end' : 'items-start'}`}
          >
            {/* 送信者名と時刻。配置と色だけでは誰の発言か伝わらないため必ず出す */}
            <p className="text-xs text-muted-foreground">
              {isMine ? '自分' : message.sender_name}・
              {formatDateTime(message.created_at)}
            </p>

            <div
              className={`max-w-[85%] rounded-lg px-4 py-3 ${
                isMine ? 'bg-primary/10' : 'border bg-card'
              }`}
            >
              <p className="whitespace-pre-wrap break-words text-sm leading-relaxed">
                {message.body}
              </p>

              {/* 伏せ字だけを見せると「なぜ消えたのか」が分からない。
                  理由（安全な取引のため）と、代わりにどうすればよいか（この画面を使う）を
                  セットで示す。マスキングはユーザーにとって不便な機能なので、
                  説明が無いと「使いにくいサービス」としか受け取られない */}
              {message.masked && (
                <p className="mt-2 flex items-start gap-1.5 border-t pt-2 text-xs text-muted-foreground">
                  <ShieldAlert
                    className="mt-0.5 size-3.5 flex-none"
                    aria-hidden="true"
                  />
                  <span>
                    連絡先とみられる情報を伏せました。安全な取引のため、やり取りはこの画面で行ってください。
                  </span>
                </p>
              )}
            </div>
          </li>
        )
      })}
    </ul>
  )
}
