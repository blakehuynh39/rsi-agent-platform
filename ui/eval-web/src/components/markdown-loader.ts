import type { ComponentType } from "react";
import type { MarkdownProps } from "@/components/Markdown";

let markdownModulePromise:
  | Promise<{ default: ComponentType<MarkdownProps> }>
  | null = null;

export function loadMarkdown() {
  markdownModulePromise ??= import("@/components/Markdown").then(
    ({ Markdown }) => ({ default: Markdown }),
  );
  return markdownModulePromise;
}

export function preloadMarkdown() {
  void loadMarkdown();
}
