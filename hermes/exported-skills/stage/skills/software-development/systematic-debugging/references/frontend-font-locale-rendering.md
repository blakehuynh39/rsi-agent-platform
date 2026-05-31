# Frontend Font / Locale Rendering Investigation

When user reports that text in a specific language renders incorrectly (misaligned/overlapping/missing diacritics, tofu glyphs, etc.), follow this investigation pattern.

## Quick Diagnosis Checklist

1. **Identify the locale** — what language is affected? (e.g., Vietnamese `vi`, Thai `th`)
2. **Find font configuration** — what custom fonts are loaded? (`@font-face`, `next/font`, CSS variables)
3. **Check font glyph coverage** — does the custom font support the target language's character set?
4. **Verify locale infrastructure** — does the app set `html[lang]`? How is locale detected?
5. **Propose a CSS-only fix** — no backend changes needed for font rendering.

## Step-by-Step Workflow

### 1. Clone the frontend repo

```bash
git clone --depth 1 --single-branch --branch develop https://github.com/owner/repo.git
```

### 2. Find font-related files

```bash
# Find all font config files
search_files pattern="*font*" target="files"
search_files pattern="@font-face|font-family|Bureau|sans-serif" target="content"
```

Key files to inspect:
- `apps/*/src/styles/fonts.css` — CSS font variable declarations (`--font-sans`, `--font-display`)
- `apps/*/src/styles/globals.css` — `@font-face` declarations loading custom .woff2 files
- `apps/*/lib/fonts.ts` — `next/font` or programmatic font loading
- `apps/*/public/fonts/` — actual font files (.woff2)

### 3. Trace the font stack

Custom fonts appear first in the CSS `font-family` stack. Example from Numo:

```css
@font-face {
  font-family: "Bureau Sans";
  src: url("/fonts/bureau-sans.woff2") format("woff2");
}

--font-sans: "Bureau Sans", system-ui, -apple-system, sans-serif;
```

When `Bureau Sans` lacks Vietnamese diacritics, the browser *should* fall back to `system-ui` — but sometimes the custom font has partial glyph coverage that renders incorrectly instead of triggering the fallback.

### 4. Check locale/i18n infrastructure

The `<html>` tag's `lang` attribute is the key enabler for locale-specific CSS:

```tsx
// __root.tsx or layout
<html lang={i18n.locale}>
```

Verify the target locale is in the supported list:

```ts
// helpers/i18n.ts
const messagesByLocale = { en, hi, bn, ta, te, vi, fil };
```

Also check feature flags that gate locale visibility:
- `VITE_VIETNAMESE_ENABLED=true`
- `VITE_FILIPINO_ENABLED=true`

### 5. Propose a fix (CSS-only)

Two approaches, both pure CSS:

**Option A — Locale-targeted CSS override** (simplest):

```css
html[lang="vi"] {
  --font-sans: system-ui, -apple-system, BlinkMacSystemFont,
    "Segoe UI", Helvetica, Arial, sans-serif;
  --font-display: Georgia, "Times New Roman", serif;
  --font-heading: system-ui, -apple-system, BlinkMacSystemFont,
    "Segoe UI", Helvetica, Arial, sans-serif;
  --font-text: system-ui, -apple-system, BlinkMacSystemFont,
    "Segoe UI", Helvetica, Arial, sans-serif;
}
```

Place this in `fonts.css` after the default `@theme` block. Works because `html[lang="vi"]` has higher specificity than `:root` / `@theme`.

**Option B — `unicode-range` on `@font-face`** (more robust):

```css
@font-face {
  font-family: "Bureau Sans";
  src: url("/fonts/bureau-sans.woff2") format("woff2");
  font-weight: 100 900;
  font-display: swap;
  unicode-range: U+0000-00FF, U+0131, U+0152-0153, U+02BB-02BC,
    U+02C6, U+02DA, U+02DC, U+0304, U+0308, U+0329,
    U+2000-206F, U+20AC, U+2122, U+2191, U+2193,
    U+2212, U+2215, U+FEFF, U+FFFD;
}
```

This restricts the custom font to Basic Latin + Latin Extended-A. For Vietnamese (which uses Latin Extended Additional characters like U+1EA0–U+1EF9), the browser will automatically use the next font in the stack (system-ui).

### 6. Identify all apps that need the fix

In a monorepo, font config may exist in multiple apps:
- `apps/web/` — main SPA
- `apps/landing/` — marketing/landing page
- `apps/admin/` — internal admin dashboard

Check each app's font setup; locale-specific overrides may need to be applied to all.

## Common Pitfalls

- **Partial glyph coverage**: Custom fonts may include some Vietnamese characters but render them with incorrect metrics (accent positions), so the browser never triggers the fallback stack. `unicode-range` is better than relying on font stack fallback.
- **`system-ui` vs locale-specific**: `system-ui` resolves to the OS default, which varies. On macOS it's San Francisco (good Vietnamese support), on Windows it's Segoe UI (also good). This is generally fine.
- **`next/font` local vs `@font-face`**: `next/font/local` doesn't directly support `unicode-range` in its config — you may need to switch to `@font-face` inline styles or use a raw CSS approach.
- **Don't forget `preload` links**: If the root layout has `<link rel="preload" as="font">` for the custom fonts, those will still preload even if unicode-range excludes them. Acceptable trade-off; the fonts won't be used but the preload cost is minor.
