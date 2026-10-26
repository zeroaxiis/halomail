import Link from "next/link";
import { Logo } from "@/components/logo";

export function SiteFooter() {
  return (
    <footer className="border-t border-border">
      <div className="container grid grid-cols-2 gap-8 py-12 sm:grid-cols-3 lg:grid-cols-5">
        <div className="col-span-2 sm:col-span-3 lg:col-span-1 space-y-3">
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
          title="Legal"
          links={[
            ["Privacy Policy", "/privacy"],
            ["Terms & Conditions", "/terms"],
            ["Acceptable Use", "/aup"],
            ["Data Processing", "/dpa"],
            ["Cookie Policy", "/cookie-policy"],
            ["Refund Policy", "/refund-policy"],
          ]}
        />
        <FooterCol
          title="Company"
          links={[
            ["GitHub", "https://github.com/zeroaxiis/halomail"],
            ["Contact Us", "mailto:support@zeroaxiis.tech"],
          ]}
        />
      </div>
      <div className="border-t border-border">
        <div className="container flex flex-col sm:flex-row gap-2 sm:gap-0 sm:h-12 py-4 sm:py-0 items-center justify-between text-xs text-muted-foreground text-center sm:text-left">
          <span>© {new Date().getFullYear()} zeroaxiis. All rights reserved.</span>
          <span className="hidden sm:inline-block italic">Everything starts from zero.</span>
          <span>Build from zero</span>
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
