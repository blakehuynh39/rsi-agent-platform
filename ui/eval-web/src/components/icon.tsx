import type { ReactNode } from "react";

export type IconName =
  | "alert"
  | "arrowDown"
  | "branch"
  | "brain"
  | "chat"
  | "chevron"
  | "circleDot"
  | "close"
  | "database"
  | "file"
  | "folder"
  | "hash"
  | "play"
  | "refresh"
  | "search"
  | "sliders"
  | "spark"
  | "terminal"
  | "wrench";

const paths: Record<IconName, ReactNode> = {
  alert: <><path d="M12 8v5" /><path d="M12 17h.01" /><path d="M10.3 4.4 2.9 17.2A2 2 0 0 0 4.6 20h14.8a2 2 0 0 0 1.7-2.8L13.7 4.4a2 2 0 0 0-3.4 0Z" /></>,
  arrowDown: <><path d="M12 5v14" /><path d="m6 13 6 6 6-6" /></>,
  branch: <><path d="M6 7h12" /><path d="M6 12h8" /><path d="M6 17h12" /><path d="M3 7h.01" /><path d="M3 12h.01" /><path d="M3 17h.01" /></>,
  brain: <><path d="M9 5a3 3 0 0 0-3 3v1.2A3.5 3.5 0 0 0 4 12.4 3.6 3.6 0 0 0 7.6 16H9" /><path d="M15 5a3 3 0 0 1 3 3v1.2a3.5 3.5 0 0 1 2 3.2A3.6 3.6 0 0 1 16.4 16H15" /><path d="M9 5v14" /><path d="M15 5v14" /><path d="M9 11h6" /></>,
  chat: <path d="M4 5.5A2.5 2.5 0 0 1 6.5 3h11A2.5 2.5 0 0 1 20 5.5v7A2.5 2.5 0 0 1 17.5 15H9l-5 4v-4.5A2.5 2.5 0 0 1 4 12.5Z" />,
  chevron: <path d="m9 6 6 6-6 6" />,
  circleDot: <><circle cx="12" cy="12" r="8" /><circle cx="12" cy="12" r="2.5" /></>,
  close: <><path d="m6 6 12 12" /><path d="M18 6 6 18" /></>,
  database: <><ellipse cx="12" cy="5" rx="7" ry="3" /><path d="M5 5v6c0 1.7 3.1 3 7 3s7-1.3 7-3V5" /><path d="M5 11v6c0 1.7 3.1 3 7 3s7-1.3 7-3v-6" /></>,
  file: <><path d="M6 3h8l4 4v14H6Z" /><path d="M14 3v5h5" /></>,
  folder: <path d="M3 6.5A2.5 2.5 0 0 1 5.5 4H10l2 2h6.5A2.5 2.5 0 0 1 21 8.5v8a2.5 2.5 0 0 1-2.5 2.5h-13A2.5 2.5 0 0 1 3 16.5Z" />,
  hash: <><path d="M10 3 8 21" /><path d="m16 3-2 18" /><path d="M4 9h17" /><path d="M3 15h17" /></>,
  play: <path d="m8 5 11 7-11 7Z" />,
  refresh: <><path d="M20 12a8 8 0 0 1-13.5 5.8" /><path d="M4 12A8 8 0 0 1 17.5 6.2" /><path d="M17.5 3.5v2.7h-2.7" /><path d="M6.5 20.5v-2.7h2.7" /></>,
  search: <><circle cx="11" cy="11" r="7" /><path d="m20 20-4-4" /></>,
  sliders: <><path d="M4 7h10" /><path d="M18 7h2" /><path d="M16 5v4" /><path d="M4 12h3" /><path d="M11 12h9" /><path d="M9 10v4" /><path d="M4 17h12" /><path d="M18 15v4" /></>,
  spark: <><path d="m12 3 1.7 5.1L19 10l-5.3 1.9L12 17l-1.7-5.1L5 10l5.3-1.9Z" /><path d="m18 15 .7 2.1L21 18l-2.3.9L18 21l-.7-2.1L15 18l2.3-.9Z" /></>,
  terminal: <><path d="M4 5h16v14H4Z" /><path d="m7 9 3 3-3 3" /><path d="M12 15h5" /></>,
  wrench: <><path d="M14.7 6.3a4 4 0 0 0 4.9 4.9l-8.3 8.3a2 2 0 0 1-2.8 0l-4-4a2 2 0 0 1 0-2.8Z" /><path d="m7 17 2-2" /></>
};

export function Icon(props: { name: IconName; className?: string }) {
  return (
    <svg className={props.className ? `icon ${props.className}` : "icon"} viewBox="0 0 24 24" aria-hidden="true">
      {paths[props.name]}
    </svg>
  );
}
