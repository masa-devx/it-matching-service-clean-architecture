---
description: 現在のブランチからPRを作成する（テンプレ準拠・Close #N 自動付与）
argument-hint: [補足があれば記載（省略可）]
allowed-tools: Bash(git *), Bash(gh pr *)
---

現在のブランチから main への Pull Request を作成してください。手順：

1. `git branch --show-current` と `git status` で現在ブランチ・未コミット変更を確認する
   - 未コミットの変更がある場合は「先にコミットしてください」と伝えて中断する
   - main ブランチにいる場合は中断する
2. ブランチ名（`phase{N}/{Issue番号}-{内容}` / `fix/{Issue番号}-{内容}`）から **Issue番号を抽出**する
3. `git log origin/main..HEAD --oneline` でこのPRに含まれるコミットを確認する
4. まだ push していなければ `git push -u origin <branch>` を実行する
5. `.github/pull_request_template.md` の構成に沿ってPR本文を作成する：
   - 概要に `Close #<Issue番号>` を必ず含める
   - 課題・対応内容は Issue の内容とコミット差分から具体的に書く
   - 動作確認セクションは実際に実行した検証（go vet / go build / npm run build 等）のみチェックを付ける
6. `gh pr create --base main` でPRを作成し、URLを報告する
7. 最後に「diffと本文を確認し、対応内容を自分の言葉に直してからセルフマージしてください」と案内する

追加の指示: $ARGUMENTS
