import { Button } from "@/components/ui/button";

export default function Home() {
  return (
    <main className="flex flex-1 flex-col items-center justify-center gap-4">
      <h1 className="text-3xl font-bold tracking-tight">Tsunagu Works</h1>
      <p className="text-muted-foreground">
        企業とIT人材をつなぐビジネスマッチング
      </p>
      <Button>はじめる</Button>
    </main>
  );
}
