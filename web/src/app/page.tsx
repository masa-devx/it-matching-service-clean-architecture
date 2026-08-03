import Link from "next/link";
import { Button } from "@/components/ui/button";
import { LogoutButton } from "@/components/LogoutButton";

export default function Home() {
  return (
    <main className="flex flex-1 flex-col items-center justify-center gap-4">
      <h1 className="text-3xl font-bold tracking-tight">Tsunagu Works</h1>
      <p className="text-muted-foreground">
        企業とIT人材をつなぐビジネスマッチング
      </p>
      <div className="flex gap-3">
        <Button asChild className="h-11">
          <Link href="/login">ログイン</Link>
        </Button>
        {/* 暫定配置: #8 で (main) ヘッダーに移し、ログイン時のみ表示にする */}
        <LogoutButton />
      </div>
    </main>
  );
}
