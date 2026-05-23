import { lazy, Suspense } from "react";
import { Spinner } from "@nous-research/ui/ui/components/spinner";
import type { MarkdownProps } from "@/components/Markdown";
import { loadMarkdown } from "@/components/markdown-loader";

const MarkdownImpl = lazy(loadMarkdown);

export function LazyMarkdown(props: MarkdownProps) {
  return (
    <Suspense fallback={<MarkdownFallback streaming={props.streaming} />}>
      <MarkdownImpl {...props} />
    </Suspense>
  );
}

function MarkdownFallback({ streaming }: { streaming?: boolean }) {
  return (
    <div
      className="flex min-h-10 items-center gap-2 text-sm text-muted-foreground"
      aria-busy="true"
    >
      <Spinner />
      <span>{streaming ? "Rendering stream..." : "Rendering markdown..."}</span>
    </div>
  );
}
