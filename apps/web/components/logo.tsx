import { cn } from "@/lib/utils";

/**
 * The HaloMail mark: a celestial floating halo ring hovering over a precision
 * geometric envelope fold. Crisp monochrome vector — zero neon colors, zero gradients,
 * zero glowing effects.
 */
export function LogoMark({
  className,
  monochrome = true,
}: {
  className?: string;
  monochrome?: boolean;
}) {
  void monochrome;
  return (
    <svg
      viewBox="6 0 52 54"
      fill="none"
      role="img"
      aria-label="HaloMail"
      className={cn("size-7 shrink-0 text-foreground", className)}
    >
      <title>HaloMail</title>
      {/* Floating celestial halo ring */}
      <ellipse
        cx="32"
        cy="9"
        rx="17"
        ry="5.5"
        stroke="currentColor"
        strokeWidth="3.2"
        strokeLinecap="round"
      />

      {/* Modern geometric envelope body */}
      <rect
        x="9"
        y="19.5"
        width="46"
        height="31"
        rx="5"
        stroke="currentColor"
        strokeWidth="3"
        fill="none"
      />

      {/* Envelope fold flap */}
      <path
        d="M10 21 L32 36 L54 21"
        stroke="currentColor"
        strokeWidth="3"
        strokeLinecap="round"
        strokeLinejoin="round"
        fill="none"
      />
    </svg>
  );
}

export function Logo({
  className,
  mark = true,
  monochrome = true,
}: {
  className?: string;
  mark?: boolean;
  monochrome?: boolean;
}) {
  void monochrome;
  return (
    <span
      className={cn(
        "inline-flex items-center gap-2.5 font-semibold tracking-tight leading-none text-foreground",
        className
      )}
    >
      {mark && <LogoMark className="size-7 shrink-0" />}
      <span>HaloMail</span>
    </span>
  );
}
