import Link from "next/link";

// 認証系画面（login / signup）共通の中央寄せレイアウト
export default function AuthLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-8 px-4 py-10">
      <Link href="/" className="text-2xl font-bold tracking-tight text-primary">
        Tsunagu Works
      </Link>
      {children}
    </div>
  );
}
