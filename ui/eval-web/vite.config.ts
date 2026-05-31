import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import path from "path";

const API_TARGET = process.env.RSI_PLATFORM_API_URL;

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
    // When @nous-research/ui is symlinked via `file:../../design-language`,
    // Node's module resolution would pick up shared deps from
    // design-language/node_modules/*, giving us two copies + breaking
    // hooks (useRef-of-null), webgl contexts, etc. Force everything that
    // exists in BOTH places to use the dashboard's copy.
    //
    // Don't list packages here that only exist in the DS (nanostores,
    // @nanostores/react) — Vite dedupe errors out when it can't find
    // them at the project root.
    dedupe: [
      "react",
      "react-dom",
      "@react-three/fiber",
      "@observablehq/plot",
      "three",
      "leva",
      "gsap",
    ],
  },
  build: {
    outDir: "../../internal/reviewui/dist",
    emptyOutDir: true,
  },
  server: API_TARGET
    ? {
        proxy: {
          "/api": {
            target: API_TARGET,
            ws: true,
            changeOrigin: true,
          },
        },
      }
    : undefined,
});
