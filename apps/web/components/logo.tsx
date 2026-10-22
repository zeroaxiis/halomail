import { cn } from "@/lib/utils";

/**
 * The HaloMail mark: three orbital rings with node points, wrapped around an
 * HM monogram. Drawn inline so it stays crisp at favicon size and picks up the
 * brand gradient in both themes.
 */
export function LogoMark({
  className,
  monochrome = false,
}: {
  className?: string;
  monochrome?: boolean;
}) {
  const color = monochrome ? "currentColor" : "url(#halomail-mark)";
  return (
    <svg
      viewBox="0 0 64 64"
      fill="none"
      role="img"
      aria-label="HaloMail"
      className={cn("size-6", className)}
    >
      <title>HaloMail</title>
      {!monochrome && (
        <defs>
          <linearGradient id="halomail-mark" x1="8" y1="6" x2="56" y2="58" gradientUnits="userSpaceOnUse">
            <stop offset="0%" stopColor="#7c4dff" />
            <stop offset="55%" stopColor="#a855f7" />
            <stop offset="100%" stopColor="#e879f9" />
          </linearGradient>
        </defs>
      )}

      <g stroke={color} strokeWidth="2.5" fill="none">
        {/* Three orbits, each rotated a third of a turn. */}
        <ellipse cx="32" cy="32" rx="25" ry="12" transform="rotate(0 32 32)" />
        <ellipse cx="32" cy="32" rx="25" ry="12" transform="rotate(60 32 32)" />
        <ellipse cx="32" cy="32" rx="25" ry="12" transform="rotate(120 32 32)" />
      </g>

      <g fill={color}>
        {/* Node points where the orbits cross their widest span. */}
        <circle cx="57" cy="32" r="3" />
        <circle cx="7" cy="32" r="3" />
        <circle cx="19.5" cy="10.4" r="3" />
        <circle cx="44.5" cy="53.6" r="3" />
        <circle cx="44.5" cy="10.4" r="3" />
        <circle cx="19.5" cy="53.6" r="3" />
      </g>

      {/* HM monogram. */}
      <text
        x="32"
        y="32"
        textAnchor="middle"
        dominantBaseline="central"
        fill={color}
        fontSize="19"
        fontWeight="700"
        fontFamily="var(--font-geist-sans), system-ui, sans-serif"
        letterSpacing="-0.5"
      >
        HM
      </text>
    </svg>
  );
}

export function Logo({
  className,
  mark = true,
  monochrome = false,
}: {
  className?: string;
  mark?: boolean;
  monochrome?: boolean;
}) {
  return (
    <span className={cn("inline-flex items-center gap-2 font-semibold tracking-tight", className)}>
      {mark && (
        <LogoMark
          monochrome={monochrome}
          className={cn(
            "size-6 shrink-0",
            !monochrome && "dark:drop-shadow-[0_0_6px_rgba(168,85,247,0.55)]",
          )}
        />
      )}
      HaloMail
    </span>
  );
}
