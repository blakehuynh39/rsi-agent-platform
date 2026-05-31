import {
  cloneElement,
  Fragment,
  isValidElement,
  useMemo,
  type ReactElement,
  type ReactNode,
} from "react";
import ReactMarkdown, { type Components } from "react-markdown";
import remarkGfm from "remark-gfm";
import { cn } from "@/lib/utils";

/**
 * Dashboard-themed markdown renderer.
 *
 * `react-markdown` handles CommonMark parsing safely, while `remark-gfm`
 * adds GitHub-flavored markdown: tables, task lists, strikethrough, footnotes,
 * and autolink literals. Element overrides keep the output aligned with the
 * dashboard's dense, mono-forward visual system instead of browser defaults.
 */
export interface MarkdownProps {
  content: string;
  highlightTerms?: string[];
  streaming?: boolean;
}

export function Markdown({
  content,
  highlightTerms,
  streaming,
}: MarkdownProps) {
  const terms = useMemo(
    () =>
      (highlightTerms ?? [])
        .map((term) => term.trim())
        .filter((term) => term.length > 0),
    [highlightTerms],
  );
  const components = useMemo(() => createComponents(terms), [terms]);

  return (
    <div className="markdown-body text-sm text-foreground leading-relaxed">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={components}
        skipHtml
      >
        {content}
      </ReactMarkdown>
      {streaming && <StreamingCaret />}
    </div>
  );
}

function createComponents(highlightTerms: string[]): Components {
  const render = (children: ReactNode) =>
    highlightReactNode(children, highlightTerms);

  return {
    h1: ({ children }) => (
      <h1 className="mb-3 mt-0 font-expanded text-lg font-bold uppercase tracking-normal text-foreground">
        {render(children)}
      </h1>
    ),
    h2: ({ children }) => (
      <h2 className="mb-2 mt-5 border-b border-border pb-1 font-expanded text-base font-bold uppercase tracking-normal text-foreground">
        {render(children)}
      </h2>
    ),
    h3: ({ children }) => (
      <h3 className="mb-1.5 mt-4 font-mono-ui text-sm font-bold text-foreground">
        {render(children)}
      </h3>
    ),
    h4: ({ children }) => (
      <h4 className="mb-1 mt-3 font-mono-ui text-xs font-bold uppercase text-muted-foreground">
        {render(children)}
      </h4>
    ),
    h5: ({ children }) => (
      <h5 className="mb-1 mt-3 font-mono-ui text-xs font-semibold text-muted-foreground">
        {render(children)}
      </h5>
    ),
    h6: ({ children }) => (
      <h6 className="mb-1 mt-3 font-mono-ui text-[0.7rem] font-semibold uppercase text-muted-foreground">
        {render(children)}
      </h6>
    ),
    p: ({ children }) => (
      <p className="my-2 text-sm leading-relaxed text-foreground/90">
        {render(children)}
      </p>
    ),
    a: ({ children, href, ...props }) => (
      <a
        {...props}
        href={href}
        target={isExternalHref(href) ? "_blank" : undefined}
        rel={isExternalHref(href) ? "noreferrer" : undefined}
        className="text-primary underline decoration-primary/30 underline-offset-2 transition-colors hover:decoration-primary/70"
      >
        {render(children)}
      </a>
    ),
    blockquote: ({ children }) => (
      <blockquote className="my-3 border-l-2 border-primary/45 bg-muted/25 px-3 py-2 text-muted-foreground">
        {render(children)}
      </blockquote>
    ),
    ul: ({ children, className }) => (
      <ul
        className={cn(
          "my-2 list-disc space-y-1 pl-5 marker:text-muted-foreground",
          className?.includes("contains-task-list") && "list-none pl-0",
        )}
      >
        {children}
      </ul>
    ),
    ol: ({ children }) => (
      <ol className="my-2 list-decimal space-y-1 pl-5 marker:text-muted-foreground">
        {children}
      </ol>
    ),
    li: ({ children, className, ...props }) => (
      <li
        {...props}
        className={cn(
          "pl-1 text-sm leading-relaxed text-foreground/90",
          className?.includes("task-list-item") &&
            "flex list-none items-start gap-2 pl-0",
        )}
      >
        {render(children)}
      </li>
    ),
    input: ({ checked, type }) => {
      if (type !== "checkbox") {
        return null;
      }
      return (
        <input
          type="checkbox"
          checked={Boolean(checked)}
          readOnly
          aria-label={checked ? "Completed" : "Not completed"}
          className="mt-[0.28rem] h-3.5 w-3.5 shrink-0 accent-primary"
        />
      );
    },
    strong: ({ children }) => (
      <strong className="font-semibold text-foreground">{render(children)}</strong>
    ),
    em: ({ children }) => (
      <em className="text-foreground/90">{render(children)}</em>
    ),
    del: ({ children }) => (
      <del className="text-muted-foreground decoration-muted-foreground/70">
        {render(children)}
      </del>
    ),
    hr: () => <hr className="my-4 border-border" />,
    pre: ({ children }) => (
      <pre className="my-3 overflow-x-auto border border-border bg-secondary/55 px-3 py-2.5 font-mono-ui text-xs leading-relaxed text-foreground">
        {children}
      </pre>
    ),
    code: ({ children, className }) => (
      <code
        className={cn(
          "font-mono-ui",
          className
            ? "bg-transparent text-inherit"
            : "bg-secondary/60 px-1.5 py-0.5 text-xs text-primary/90",
          className,
        )}
      >
        {children}
      </code>
    ),
    table: ({ children }) => (
      <div className="my-3 overflow-x-auto border border-border bg-background/45">
        <table className="w-full border-collapse font-mono-ui text-xs">
          {children}
        </table>
      </div>
    ),
    thead: ({ children }) => (
      <thead className="border-b border-border bg-muted/35 text-muted-foreground">
        {children}
      </thead>
    ),
    tbody: ({ children }) => <tbody>{children}</tbody>,
    tr: ({ children }) => (
      <tr className="border-b border-border/70 last:border-b-0">{children}</tr>
    ),
    th: ({ children, align }) => (
      <th
        align={align}
        className="px-2.5 py-2 text-left font-bold uppercase tracking-normal text-muted-foreground"
      >
        {render(children)}
      </th>
    ),
    td: ({ children, align }) => (
      <td align={align} className="px-2.5 py-2 align-top text-foreground/90">
        {render(children)}
      </td>
    ),
    img: ({ alt, src }) => {
      if (!src || !isSafeImageSrc(src)) {
        return (
          <span className="my-3 inline-block border border-border bg-muted/30 px-2 py-1 text-xs text-muted-foreground">
            [Image: {alt || "blocked"}]
          </span>
        );
      }
      return (
        <img
          alt={alt ?? ""}
          src={src}
          className="my-3 max-h-[28rem] max-w-full border border-border object-contain"
        />
      );
    },
    sup: ({ children }) => (
      <sup className="font-mono-ui text-[0.65rem] text-primary">
        {render(children)}
      </sup>
    ),
    section: ({ children, className }) => (
      <section
        className={cn(
          "mt-5 border-t border-border pt-3 text-xs text-muted-foreground",
          className,
        )}
      >
        {render(children)}
      </section>
    ),
  };
}

function StreamingCaret() {
  return (
    <span
      aria-hidden
      className="inline-block h-[1em] w-[0.5em] align-[-0.15em] bg-foreground/50 animate-pulse"
    />
  );
}

function highlightReactNode(node: ReactNode, terms: string[]): ReactNode {
  if (terms.length === 0) {
    return node;
  }
  if (typeof node === "string" || typeof node === "number") {
    return <HighlightedText text={String(node)} terms={terms} />;
  }
  if (Array.isArray(node)) {
    return node.map((child, index) => (
      <Fragment key={index}>{highlightReactNode(child, terms)}</Fragment>
    ));
  }
  if (isValidElement(node)) {
    if (node.type === "code" || node.type === "pre") {
      return node;
    }
    const element = node as ReactElement<{ children?: ReactNode }>;
    if (!("children" in element.props)) {
      return node;
    }
    return cloneElement(element, {
      children: highlightReactNode(element.props.children, terms),
    });
  }
  return node;
}

function HighlightedText({ text, terms }: { text: string; terms: string[] }) {
  const escaped = terms.map((term) => term.replace(/[.*+?^${}()|[\]\\]/g, "\\$&"));
  const splitRegex = new RegExp(`(${escaped.join("|")})`, "gi");
  const exactRegex = new RegExp(`^(?:${escaped.join("|")})$`, "i");
  const parts = text.split(splitRegex);

  return (
    <>
      {parts.map((part, index) =>
        part && exactRegex.test(part) ? (
          <mark key={index} className="bg-warning/30 px-0.5 text-warning">
            {part}
          </mark>
        ) : (
          <span key={index}>{part}</span>
        ),
      )}
    </>
  );
}

function isSafeImageSrc(src: string): boolean {
  return !(/^[a-z][a-z0-9+.-]*:/i.test(src));
}

function isExternalHref(href: string | undefined): boolean {
  return Boolean(href && /^[a-z][a-z0-9+.-]*:/i.test(href));
}
