/**
 * Crisp RSI variant of the Hermes dashboard backdrop.
 *
 * Upstream Hermes mirrors the Nous DS overlay stack, including a filler JPEG
 * and SVG turbulence layer. That looks intentionally analog, but it makes RSI
 * traces and final answers hard to scan. Keep the theme colors and optional
 * theme asset hook, while making the default surface flat and text-first.
 */
export function Backdrop() {
  return (
    <>
      <div
        aria-hidden
        className="pointer-events-none fixed inset-0 z-[1]"
        style={{
          backgroundColor: "var(--background-base)",
        }}
      />

      <div
        aria-hidden
        className="pointer-events-none fixed inset-0 z-[2]"
        style={
          {
            opacity: "var(--component-backdrop-asset-opacity, 0)",
            backgroundImage: "var(--theme-asset-bg)",
            backgroundSize: "var(--component-backdrop-background-size, cover)",
            backgroundPosition:
              "var(--component-backdrop-background-position, center)",
          } as unknown as React.CSSProperties
        }
      />

      <div
        aria-hidden
        className="pointer-events-none fixed inset-0 z-[99]"
        style={{
          background:
            "radial-gradient(ellipse at 0% 0%, transparent 60%, var(--warm-glow) 100%)",
          mixBlendMode: "lighten",
          opacity: "var(--component-backdrop-vignette-opacity, 0.08)",
        }}
      />
    </>
  );
}
