'use client'

import type {
  TsunaguWorksProject,
  TsunaguWorksProjectStatus,
} from '@repo/api-client/company/generated/models'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import { Button } from '@/components/ui/button'
import type { ProjectStatusAction } from '@/external/handler/project'

import { changeProjectStatusAction } from '../../actions/status'
import { companyProjectKeys } from '../../queries/companyProjects'

// 状態ごとに出す操作の定義（shared/domain の遷移表と同じ知識の UI 側の写し）。
// UI は「普段させない」ためのガイドで、正しさの保証はサーバーの遷移表＋条件付きUPDATE。
// タブ2枚などで食い違ったときは 409 がトーストで返る。
// confirm があるものは外から見える変化（公開・終了）なので確認ダイアログを挟む
const successMessage: Record<ProjectStatusAction, string> = {
  publish: 'を公開しました',
  unpublish: 'を非公開にしました',
  close: 'の募集を終了しました',
}

const actionsByStatus: Record<
  TsunaguWorksProjectStatus,
  Array<{
    action: ProjectStatusAction
    label: string
    confirm?: { title: string; description: string }
  }>
> = {
  draft: [
    {
      action: 'publish',
      label: '公開する',
      confirm: {
        title: 'この案件を公開しますか？',
        description: '公開すると人材の検索結果に表示され、応募を受け付けます。',
      },
    },
  ],
  published: [
    { action: 'unpublish', label: '非公開にする' },
    {
      action: 'close',
      label: '募集終了',
      confirm: {
        title: 'この案件の募集を終了しますか？',
        description: '検索結果から外れます（あとから再公開もできます）。',
      },
    },
  ],
  closed: [
    {
      action: 'publish',
      label: '再公開する',
      confirm: {
        title: 'この案件を再公開しますか？',
        description: '再び人材の検索結果に表示され、応募を受け付けます。',
      },
    },
  ],
}

export function ProjectStatusActions({
  project,
}: {
  project: TsunaguWorksProject
}) {
  const queryClient = useQueryClient()

  const mutation = useMutation({
    mutationFn: async (action: ProjectStatusAction) => {
      const result = await changeProjectStatusAction(project.id, action)
      if (!result.ok) {
        throw new Error(result.error)
      }
      return result.data
    },
    // 一覧・詳細をまとめて無効化（前方一致）→ 表示中の useQuery が自動で再取得する。
    // 成功トーストは対象名入り（一覧に同種ボタンが並ぶ画面では「どれに効いたか」を必ず伝える）
    onSuccess: (_updated, action) => {
      toast.success(`「${project.title}」${successMessage[action]}`)
      void queryClient.invalidateQueries({ queryKey: companyProjectKeys.all })
    },
    // 409（遷移競合）は Go が「可能な遷移先つき」の文言を返すのでそのまま見せる
    onError: (error) => {
      toast.error(error.message)
    },
  })

  return (
    <div className="flex gap-2">
      {actionsByStatus[project.status].map(({ action, label, confirm }) =>
        confirm ? (
          <AlertDialog key={action}>
            <AlertDialogTrigger asChild>
              <Button variant="outline" size="sm" disabled={mutation.isPending}>
                {label}
              </Button>
            </AlertDialogTrigger>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>{confirm.title}</AlertDialogTitle>
                <AlertDialogDescription>
                  「{project.title}」 — {confirm.description}
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel>キャンセル</AlertDialogCancel>
                <AlertDialogAction onClick={() => mutation.mutate(action)}>
                  {label}
                </AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        ) : (
          <Button
            key={action}
            variant="outline"
            size="sm"
            disabled={mutation.isPending}
            onClick={() => mutation.mutate(action)}
          >
            {label}
          </Button>
        ),
      )}
    </div>
  )
}
