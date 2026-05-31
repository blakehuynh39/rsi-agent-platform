import {
  useCallback,
  useEffect,
  useLayoutEffect,
  memo,
  useMemo,
  useRef,
  useState,
  type MouseEvent,
} from "react";
import { FileText, RefreshCw, Search } from "lucide-react";
import { Spinner } from "@nous-research/ui/ui/components/spinner";
import { LazyMarkdown } from "@/components/LazyMarkdown";
import { preloadMarkdown } from "@/components/markdown-loader";
import { usePageHeader } from "@/contexts/usePageHeader";
import {
  api,
  type CompanyWikiMarkdownRead,
  type CompanyWikiSearchResult,
} from "@/lib/api";
import { cn, stripFrontmatter } from "@/lib/utils";
import { PluginSlot } from "@/plugins";

const INDEX_PATH = "index.md";
const WIKI_INDEX_FALLBACK_SECTION = "pages";
const WIKI_INDEX_NAV_PREVIEW_LIMIT = 6;
const WIKI_INDEX_SECTION_PREVIEW_LIMIT = 12;

const DS_BUTTON_CN = cn(
  "inline-flex items-center justify-center gap-2",
  "border border-current/25 bg-transparent px-3 py-2",
  "font-mono-ui text-xs font-bold text-midground",
  "shadow-[inset_-1px_-1px_0_0_#00000080,inset_1px_1px_0_0_#ffffff80]",
  "transition-colors hover:bg-current/10 disabled:cursor-not-allowed disabled:opacity-50",
);

export default function DocsPage() {
  const { setEnd } = usePageHeader();
  const [indexContent, setIndexContent] = useState("");
  const [currentPath, setCurrentPath] = useState(INDEX_PATH);
  const [currentContent, setCurrentContent] = useState("");
  const [results, setResults] = useState<CompanyWikiSearchResult[]>([]);
  const [loading, setLoading] = useState(true);
  const [searching, setSearching] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const fileCacheRef = useRef(new Map<string, CompanyWikiMarkdownRead>());
  const searchCacheRef = useRef(new Map<string, CompanyWikiSearchResult[]>());
  const fileAbortRef = useRef<AbortController | null>(null);
  const searchAbortRef = useRef<AbortController | null>(null);
  const indexData = useMemo(() => parseWikiIndex(indexContent), [indexContent]);
  const viewingIndex = currentPath === INDEX_PATH;

  const loadFile = useCallback(
    async (path: string, options?: { refresh?: boolean }) => {
      const normalized = normalizeWikiPath(path) || INDEX_PATH;
      const cached = options?.refresh
        ? undefined
        : fileCacheRef.current.get(normalized);
      setLoading(!cached);
      setError(null);
      if (cached) {
        setCurrentPath(cached.path || normalized);
        setCurrentContent(cached.content || "");
        if (normalized === INDEX_PATH) {
          setIndexContent(cached.content || "");
        }
        return;
      }
      fileAbortRef.current?.abort();
      const controller = new AbortController();
      fileAbortRef.current = controller;
      try {
        const read =
          normalized === INDEX_PATH
            ? await api.getCompanyWikiIndex(controller.signal)
            : await api.getCompanyWikiFile(normalized, controller.signal);
        if (controller.signal.aborted) {
          return;
        }
        fileCacheRef.current.set(normalized, read);
        setCurrentPath(read.path || normalized);
        setCurrentContent(read.content || "");
        if (normalized === INDEX_PATH) {
          setIndexContent(read.content || "");
        }
      } catch (err) {
        if (controller.signal.aborted) {
          return;
        }
        setError(err instanceof Error ? err.message : String(err));
      } finally {
        if (fileAbortRef.current === controller) {
          fileAbortRef.current = null;
          setLoading(false);
        }
      }
    },
    [],
  );

  const refresh = useCallback(() => {
    void loadFile(currentPath || INDEX_PATH, { refresh: true });
  }, [currentPath, loadFile]);

  useLayoutEffect(() => {
    setEnd(
      <button type="button" onClick={refresh} className={DS_BUTTON_CN}>
        <RefreshCw className="size-3.5" />
        Refresh wiki
      </button>,
    );
    return () => {
      setEnd(null);
    };
  }, [refresh, setEnd]);

  useEffect(() => {
    preloadMarkdown();
    queueMicrotask(() => {
      void loadFile(INDEX_PATH);
    });
    return () => {
      fileAbortRef.current?.abort();
      searchAbortRef.current?.abort();
    };
  }, [loadFile]);

  const visibleContent = useMemo(
    () => (viewingIndex ? "" : stripFrontmatter(currentContent)),
    [currentContent, viewingIndex],
  );

  const handleCurrentWikiLinkClick = useCallback(
    (event: MouseEvent<HTMLElement>) => {
      handleWikiLinkClick(event, currentPath, loadFile);
    },
    [currentPath, loadFile],
  );

  const runSearch = useCallback(async (query: string) => {
    const trimmed = query.trim();
    if (!trimmed) {
      searchAbortRef.current?.abort();
      searchAbortRef.current = null;
      setSearching(false);
      setResults([]);
      void loadFile(INDEX_PATH);
      return;
    }
    const cacheKey = trimmed.toLowerCase();
    const cached = searchCacheRef.current.get(cacheKey);
    if (cached) {
      searchAbortRef.current?.abort();
      searchAbortRef.current = null;
      setSearching(false);
      setResults(cached);
      return;
    }
    searchAbortRef.current?.abort();
    const controller = new AbortController();
    searchAbortRef.current = controller;
    setSearching(true);
    setError(null);
    try {
      const response = await api.searchCompanyWiki(
        trimmed,
        25,
        controller.signal,
      );
      if (controller.signal.aborted) {
        return;
      }
      const nextResults = response.results ?? [];
      searchCacheRef.current.set(cacheKey, nextResults);
      setResults(nextResults);
    } catch (err) {
      if (controller.signal.aborted) {
        return;
      }
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      if (searchAbortRef.current === controller) {
        searchAbortRef.current = null;
        setSearching(false);
      }
    }
  }, [loadFile]);

  return (
    <div className="flex min-h-0 w-full min-w-0 flex-1 flex-col pt-1 sm:pt-2">
      <PluginSlot name="docs:top" />

      <div className="mb-3 flex flex-col gap-2 border-b border-current/15 pb-3 sm:flex-row sm:items-center">
        <CompanyWikiSearchForm onSearch={runSearch} searching={searching} />
        <button
          type="button"
          className={DS_BUTTON_CN}
          onClick={() => {
            setResults([]);
            void loadFile(INDEX_PATH);
          }}
        >
          <FileText className="size-3.5" />
          Index
        </button>
      </div>

      <div className="grid min-h-0 flex-1 gap-3 lg:grid-cols-[minmax(16rem,22rem)_minmax(0,1fr)]">
        <aside className="min-h-0 overflow-auto border border-current/15 bg-background/40 p-3">
          <div className="mb-2 font-mono-ui text-xs font-bold uppercase text-muted-foreground">
            Company Wiki
          </div>
          {results.length > 0 ? (
            <SearchResults results={results} loadFile={loadFile} />
          ) : (
            <WikiIndexNav index={indexData} loadFile={loadFile} />
          )}
        </aside>

        <main className="min-h-0 overflow-auto border border-current/15 bg-background p-4">
          <div className="mb-3 flex min-w-0 items-center justify-between gap-3 border-b border-current/10 pb-2">
            <div className="min-w-0 truncate font-mono-ui text-xs text-muted-foreground">
              {currentPath}
            </div>
          </div>
          {loading ? (
            <div className="flex min-h-48 items-center justify-center gap-2 text-sm text-muted-foreground">
              <Spinner />
              Loading company wiki...
            </div>
          ) : error ? (
            <div className="border border-destructive/40 bg-destructive/10 p-3 text-sm text-destructive">
              {error}
            </div>
          ) : viewingIndex ? (
            <WikiIndexOverview index={indexData} loadFile={loadFile} />
          ) : (
            <article
              className="wiki-markdown max-w-5xl"
              onClick={handleCurrentWikiLinkClick}
            >
              <LazyMarkdown content={visibleContent} />
            </article>
          )}
        </main>
      </div>

      <PluginSlot name="docs:bottom" />
    </div>
  );
}

const CompanyWikiSearchForm = memo(function CompanyWikiSearchForm({
  onSearch,
  searching,
}: {
  onSearch: (query: string) => void | Promise<void>;
  searching: boolean;
}) {
  const [query, setQuery] = useState("");

  return (
    <form
      className="flex min-w-0 flex-1 items-center gap-2"
      onSubmit={(event) => {
        event.preventDefault();
        void onSearch(query);
      }}
    >
      <label className="sr-only" htmlFor="company-wiki-search">
        Search company wiki
      </label>
      <div className="relative min-w-0 flex-1">
        <Search className="pointer-events-none absolute left-3 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
        <input
          id="company-wiki-search"
          value={query}
          onChange={(event) => setQuery(event.target.value)}
          placeholder="Search company wiki..."
          className={cn(
            "h-10 w-full border border-current/25 bg-background pl-9 pr-3",
            "font-mono-ui text-sm text-foreground outline-none",
            "focus:border-primary",
          )}
        />
      </div>
      <button type="submit" className={DS_BUTTON_CN} disabled={searching}>
        {searching ? <Spinner /> : <Search className="size-3.5" />}
        Search
      </button>
    </form>
  );
});

function SearchResults({
  results,
  loadFile,
}: {
  results: CompanyWikiSearchResult[];
  loadFile: (path: string) => void | Promise<void>;
}) {
  return (
    <div className="space-y-2">
      <div className="font-mono-ui text-xs text-muted-foreground">
        {results.length} result{results.length === 1 ? "" : "s"}
      </div>
      {results.map((result) => (
        <button
          key={result.wiki_revision_id || result.path}
          type="button"
          onClick={() => void loadFile(result.path)}
          className="block w-full border border-current/15 p-2 text-left transition-colors hover:bg-current/10"
        >
          <div className="truncate text-sm font-semibold text-foreground">
            {result.title || result.slug || result.path}
          </div>
          <div className="mt-1 truncate font-mono-ui text-[0.68rem] text-muted-foreground">
            {result.path}
          </div>
          {result.snippet ? (
            <div className="mt-1 max-h-9 overflow-hidden text-xs text-muted-foreground">
              {result.snippet}
            </div>
          ) : null}
        </button>
      ))}
    </div>
  );
}

interface WikiIndexEntry {
  title: string;
  path: string;
  summary: string;
}

interface WikiIndexSection {
  title: string;
  entries: WikiIndexEntry[];
}

interface WikiIndexData {
  description: string;
  sections: WikiIndexSection[];
}

function WikiIndexNav({
  index,
  loadFile,
}: {
  index: WikiIndexData;
  loadFile: (path: string) => void | Promise<void>;
}) {
  const total = countWikiIndexEntries(index);

  if (total === 0) {
    return (
      <p className="text-sm text-muted-foreground">
        No published wiki pages yet.
      </p>
    );
  }

  return (
    <div className="space-y-4">
      {index.sections.map((section) => (
        <section key={section.title}>
          <h2 className="mb-1.5 font-mono-ui text-xs font-bold uppercase text-muted-foreground">
            {section.title}
            <span className="ml-1 font-normal normal-case">
              ({section.entries.length})
            </span>
          </h2>
          <div className="space-y-1">
            {section.entries
              .slice(0, WIKI_INDEX_NAV_PREVIEW_LIMIT)
              .map((entry) => (
                <button
                  key={entry.path}
                  type="button"
                  onClick={() => void loadFile(entry.path)}
                  className="block w-full border border-current/10 px-2 py-1.5 text-left transition-colors hover:bg-current/10"
                >
                  <span className="block truncate text-sm font-medium text-foreground">
                    {entry.title}
                  </span>
                  <span className="mt-0.5 block truncate font-mono-ui text-[0.66rem] text-muted-foreground">
                    {entry.path}
                  </span>
                </button>
              ))}
            {section.entries.length > WIKI_INDEX_NAV_PREVIEW_LIMIT ? (
              <div className="px-2 pt-1 font-mono-ui text-[0.66rem] text-muted-foreground">
                +{section.entries.length - WIKI_INDEX_NAV_PREVIEW_LIMIT} more
              </div>
            ) : null}
          </div>
        </section>
      ))}
    </div>
  );
}

function WikiIndexOverview({
  index,
  loadFile,
}: {
  index: WikiIndexData;
  loadFile: (path: string) => void | Promise<void>;
}) {
  const total = countWikiIndexEntries(index);
  const [expandedSections, setExpandedSections] = useState<Set<string>>(
    () => new Set(),
  );

  const toggleSection = useCallback((sectionTitle: string) => {
    setExpandedSections((current) => {
      const next = new Set(current);
      if (next.has(sectionTitle)) {
        next.delete(sectionTitle);
      } else {
        next.add(sectionTitle);
      }
      return next;
    });
  }, []);

  return (
    <div className="max-w-6xl">
      <div className="mb-4 border-b border-current/10 pb-3">
        <h1 className="font-expanded text-lg font-bold uppercase text-foreground">
          Company Wiki Index
        </h1>
        <p className="mt-1 max-w-3xl text-sm text-muted-foreground">
          {index.description ||
            "Browse the generated company wiki catalog and open a page to render its full markdown."}
        </p>
        <div className="mt-2 font-mono-ui text-xs text-muted-foreground">
          {total} page{total === 1 ? "" : "s"}
        </div>
      </div>

      {total === 0 ? (
        <div className="border border-current/15 bg-muted/20 p-3 text-sm text-muted-foreground">
          No published wiki pages yet.
        </div>
      ) : (
        <div className="space-y-5">
          {index.sections.map((section) => (
            <section key={section.title}>
              <div className="mb-2 flex min-w-0 items-center justify-between gap-3">
                <h2 className="font-mono-ui text-xs font-bold uppercase text-muted-foreground">
                  {section.title}
                  <span className="ml-1 font-normal normal-case">
                    ({section.entries.length})
                  </span>
                </h2>
                {section.entries.length > WIKI_INDEX_SECTION_PREVIEW_LIMIT ? (
                  <button
                    type="button"
                    onClick={() => toggleSection(section.title)}
                    className="font-mono-ui text-xs text-primary underline decoration-primary/30 underline-offset-2 hover:decoration-primary/70"
                  >
                    {expandedSections.has(section.title)
                      ? "Show fewer"
                      : `Show all ${section.entries.length}`}
                  </button>
                ) : null}
              </div>
              <div className="grid gap-2 md:grid-cols-2 xl:grid-cols-3">
                {(expandedSections.has(section.title)
                  ? section.entries
                  : section.entries.slice(0, WIKI_INDEX_SECTION_PREVIEW_LIMIT)
                ).map((entry) => (
                  <button
                    key={entry.path}
                    type="button"
                    onClick={() => void loadFile(entry.path)}
                    className="min-h-24 border border-current/15 bg-background/70 p-3 text-left transition-colors hover:bg-current/10"
                  >
                    <span className="block text-sm font-semibold text-foreground">
                      {entry.title}
                    </span>
                    <span className="mt-1 block truncate font-mono-ui text-[0.68rem] text-muted-foreground">
                      {entry.path}
                    </span>
                    {entry.summary ? (
                      <span className="mt-2 line-clamp-2 block text-xs leading-relaxed text-muted-foreground">
                        {entry.summary}
                      </span>
                    ) : null}
                  </button>
                ))}
              </div>
            </section>
          ))}
        </div>
      )}
    </div>
  );
}

function resolveWikiHref(href: string, currentPath: string): string {
  href = href.trim();
  if (!href || href.startsWith("#") || /^[a-z][a-z0-9+.-]*:/i.test(href)) {
    return "";
  }
  const withoutHash = href.split("#", 1)[0] ?? "";
  if (!withoutHash) {
    return "";
  }
  if (withoutHash.startsWith("/")) {
    return normalizeWikiPath(withoutHash);
  }
  const base = currentPath.includes("/")
    ? currentPath.slice(0, currentPath.lastIndexOf("/") + 1)
    : "";
  return normalizeWikiPath(base + withoutHash);
}

function handleWikiLinkClick(
  event: MouseEvent<HTMLElement>,
  currentPath: string,
  loadFile: (path: string) => void | Promise<void>,
) {
  const target = event.target;
  if (!(target instanceof HTMLElement)) {
    return;
  }
  const anchor = target.closest("a");
  const href = anchor?.getAttribute("href") ?? "";
  const wikiPath = resolveWikiHref(href, currentPath);
  if (!wikiPath) {
    return;
  }
  event.preventDefault();
  void loadFile(wikiPath);
}

function normalizeWikiPath(path: string): string {
  const parts: string[] = [];
  for (const part of path.replace(/\\/g, "/").split("/")) {
    if (!part || part === ".") {
      continue;
    }
    if (part === "..") {
      parts.pop();
      continue;
    }
    parts.push(part);
  }
  const normalized = parts.join("/");
  if (
    normalized === INDEX_PATH ||
    normalized === "log.md" ||
    normalized === "SCHEMA.md" ||
    normalized.startsWith("pages/") ||
    normalized.startsWith("sources/")
  ) {
    return normalized;
  }
  return "";
}

function stripIndexMetadata(content: string): string {
  return content.replace(/ `updated=[^`]+`/g, "");
}

function parseWikiIndex(content: string): WikiIndexData {
  const sections: WikiIndexSection[] = [];
  const descriptionLines: string[] = [];
  let currentSection: WikiIndexSection | null = null;

  for (const line of stripFrontmatter(content).split(/\r?\n/)) {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith("# Company Wiki Index")) {
      continue;
    }

    const heading = trimmed.match(/^##\s+(.+)$/);
    if (heading) {
      currentSection = { title: heading[1].trim(), entries: [] };
      sections.push(currentSection);
      continue;
    }

    const entry = parseWikiIndexEntry(trimmed);
    if (entry) {
      if (!currentSection) {
        currentSection = { title: WIKI_INDEX_FALLBACK_SECTION, entries: [] };
        sections.push(currentSection);
      }
      currentSection.entries.push(entry);
      continue;
    }

    if (!currentSection && !trimmed.startsWith("#")) {
      descriptionLines.push(trimmed);
    }
  }

  return {
    description: descriptionLines.join(" ").trim(),
    sections: sections.filter((section) => section.entries.length > 0),
  };
}

function parseWikiIndexEntry(line: string): WikiIndexEntry | null {
  const match = line.match(/^- \[([^\]]+)\]\(([^)]+)\)(?:\s+[—-]\s+(.+))?$/);
  if (!match) {
    return null;
  }
  const path = normalizeWikiPath(match[2]);
  if (!path || path === INDEX_PATH) {
    return null;
  }
  return {
    title: unescapeMarkdownText(match[1]),
    path,
    summary: stripIndexMetadata(match[3] ?? "").trim(),
  };
}

function unescapeMarkdownText(value: string): string {
  return value.replace(/\\([\\[\]])/g, "$1");
}

function countWikiIndexEntries(index: WikiIndexData): number {
  return index.sections.reduce(
    (total, section) => total + section.entries.length,
    0,
  );
}
