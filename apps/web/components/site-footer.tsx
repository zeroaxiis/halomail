import Link from "next/link";
import { Logo } from "@/components/logo";

export function SiteFooter() {
  return (
    <footer className="border-t border-border">
      <div className="container grid gap-8 py-12 md:grid-cols-4">
        <div className="space-y-3">
          <Logo />
          <p className="max-w-xs text-sm text-muted-foreground">
            API-first scheduling and contact forms. Minimal, premium,
            developer-focused.
          </p>
        </div>
        <FooterCol
          title="Product"
          links={[
            ["Features", "/#features"],
            ["Pricing", "/pricing"],
            ["Booking", "/book/demo"],
          ]}
        />
        <FooterCol
          title="Developers"
          links={[
            ["Docs", "/docs/index.html"],
            ["Add to your site", "/docs/integrate.html"],
            ["API keys", "/dashboard/api-keys"],
            ["Status", "#"],
          ]}
        />
        <FooterCol
          title="Company"
          links={[
            ["GitHub", "https://github.com/zeroaxiis/halomail"],
            ["Privacy", "#"],
            ["Terms", "#"],
          ]}
        />
      </div>
      <div className="border-t border-border">
        <div className="container flex h-12 items-center justify-between text-xs text-muted-foreground">
          <span>© {new Date().getFullYear()} HaloMail. MIT licensed.</span>
          <span>Built with 👍🏻</span>
        </div>
      </div>
    </footer>
  );
}

function FooterCol({
  title,
  links,
}: {
  title: string;
  links: [string, string][];
}) {
  return (
    <div className="space-y-3 text-sm">
      <div className="font-medium">{title}</div>
      <ul className="space-y-2 text-muted-foreground">
        {links.map(([label, href]) => (
          <li key={label}>
            <Link
              href={href}
              className="transition-colors hover:text-foreground"
            >
              {label}
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
}
