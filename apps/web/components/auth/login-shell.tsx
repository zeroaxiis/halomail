import Link from "next/link";
import { Logo } from "@/components/logo";
import { ThemeToggle } from "@/components/theme-toggle";
import styles from "./login.module.css";

// Official mark from zeroaxiis/zeroaxiis, components/icons/brand.tsx.
function ZeroAxiisMark() {
  return (
    <svg width="28" height="28" viewBox="0 0 100 100" fill="none" aria-hidden="true">
      <circle cx="50" cy="50" r="46" fill="#FFCC00" stroke="#1A1C20" strokeWidth="6" />
      <g stroke="#1A1C20" strokeWidth="6">
        <path d="M50 4v92" />
        <ellipse cx="50" cy="50" rx="22" ry="46" />
        <path d="M10 26Q50 10 90 26M4 50Q50 35 96 50M10 74Q50 60 90 74" strokeLinecap="round" />
      </g>
    </svg>
  );
}

const offers = [
  { title: "Scheduling", detail: "Booking pages that work around you." },
  { title: "Contact forms", detail: "Every conversation, in one inbox." },
  { title: "Email & API", detail: "Your templates. Your integrations." },
];

export function LoginShell({ children }: { children: React.ReactNode }) {
  return (
    <main className={styles.shell}>
      <header className={styles.header}>
        <Link href="/" aria-label="HaloMail home" className={styles.homeLink}>
          <Logo className="text-xl" />
        </Link>
        <div className={styles.headerActions}>
          <Link href="/register" className={styles.registerLink}>New here? <span>Create an account</span></Link>
          <ThemeToggle />
        </div>
      </header>

      <div className={styles.content}>
        <section className={styles.story} aria-labelledby="login-story-title">
          <p className={styles.kicker}>Good things begin with a connection.</p>
          <h1 id="login-story-title" className={styles.headline}>
            Everything{" "}<br />starts from <span>zero.</span>
          </h1>
          <p className={styles.intro}>A meeting booked. A message received.<br />A little more room for what you&apos;re building.</p>
          <ul className={styles.offers} aria-label="What HaloMail offers">
            {offers.map((offer) => (
              <li key={offer.title}>
                <h2>{offer.title}</h2>
                <p>{offer.detail}</p>
              </li>
            ))}
          </ul>
        </section>

        <section className={styles.formArea} aria-label="Sign in to HaloMail">
          {children}
        </section>
      </div>

      <footer className={styles.footer}>
        <div className={styles.company}>
          <a href="https://zeroaxiis.tech" target="_blank" rel="noopener noreferrer" className={styles.parentBrand} aria-label="HaloMail is a product of ZeroAxiis. Learn about us (opens in a new tab)">
            <ZeroAxiisMark />
            <span><span className={styles.productOf}>A product of</span><strong>ZeroAxiis</strong></span>
          </a>
          <a href="https://zeroaxiis.tech" target="_blank" rel="noopener noreferrer" className={styles.aboutLink}>
            <span>About us</span><span>zeroaxiis.tech</span>
          </a>
        </div>
        <nav className={styles.legal} aria-label="Legal">
          <span>© {new Date().getFullYear()} HaloMail</span>
          <Link href="/privacy">Privacy</Link>
          <Link href="/terms">Terms</Link>
        </nav>
      </footer>
    </main>
  );
}

