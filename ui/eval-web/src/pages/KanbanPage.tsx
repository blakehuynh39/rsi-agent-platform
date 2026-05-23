import { useCallback, useEffect, useMemo, useState } from "react";
import {
  Archive,
  Circle,
  CircleCheck,
  CircleDashed,
  Columns3,
  MessageSquare,
  Plus,
  RefreshCw,
  Search,
  X,
} from "lucide-react";
import { Button } from "@nous-research/ui/ui/components/button";
import { Badge } from "@nous-research/ui/ui/components/badge";
import { Spinner } from "@nous-research/ui/ui/components/spinner";
import { Input } from "@/components/ui/input";
import { api } from "@/lib/api";
import type {
  KanbanBoardSnapshot,
  KanbanProject,
  KanbanTicket,
  KanbanTicketComment,
  KanbanTicketSourceRef,
  KanbanTicketStatus,
} from "@/lib/api";
import { cn, timeAgo } from "@/lib/utils";

const STATUSES: Array<{
  status: KanbanTicketStatus;
  label: string;
  icon: typeof Circle;
}> = [
  { status: "triage", label: "Triage", icon: CircleDashed },
  { status: "todo", label: "Todo", icon: Circle },
  { status: "in_progress", label: "In progress", icon: RefreshCw },
  { status: "blocked", label: "Blocked", icon: X },
  { status: "done", label: "Done", icon: CircleCheck },
  { status: "archived", label: "Archived", icon: Archive },
];

const STATUS_LABEL = Object.fromEntries(
  STATUSES.map((item) => [item.status, item.label]),
) as Record<KanbanTicketStatus, string>;

function isStatusTransitionAllowed(
  from: KanbanTicketStatus,
  to: KanbanTicketStatus,
): boolean {
  if (from === to) return true;
  if (from === "archived") return to === "todo";
  switch (from) {
    case "triage":
      return to === "todo";
    case "todo":
      return to === "in_progress";
    case "in_progress":
      return to === "blocked" || to === "done";
    case "blocked":
      return to === "todo" || to === "in_progress";
    case "done":
      return to === "todo" || to === "archived";
    default:
      return false;
  }
}

export default function KanbanPage() {
  const [projects, setProjects] = useState<KanbanProject[]>([]);
  const [projectID, setProjectID] = useState("");
  const [snapshot, setSnapshot] = useState<KanbanBoardSnapshot | null>(null);
  const [selectedID, setSelectedID] = useState("");
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [creatingProject, setCreatingProject] = useState(false);
  const [newProjectName, setNewProjectName] = useState("");
  const [newTicketTitle, setNewTicketTitle] = useState("");
  const [newTicketDescription, setNewTicketDescription] = useState("");

  const loadProjects = useCallback(async () => {
    const response = await api.getKanbanProjects();
    setProjects(response.projects);
    setProjectID((current) => current || response.projects[0]?.id || "");
  }, []);

  const loadBoard = useCallback(async (id: string, signal?: AbortSignal) => {
    if (!id) {
      setSnapshot(null);
      return;
    }
    const board = await api.getKanbanBoard(id, signal);
    setSnapshot(board);
  }, []);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError("");
    loadProjects()
      .catch((err) => {
        if (!cancelled) setError(String(err.message || err));
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [loadProjects]);

  useEffect(() => {
    if (!projectID) return;
    const controller = new AbortController();
    setError("");
    loadBoard(projectID, controller.signal).catch((err) => {
      if (err.name !== "AbortError") setError(String(err.message || err));
    });
    return () => controller.abort();
  }, [projectID, loadBoard]);

  const latestEventID = useMemo(() => {
    if (!snapshot || snapshot.project.id !== projectID) return "";
    return snapshot.events?.at(-1)?.id ?? "";
  }, [snapshot, projectID]);
  const boardReady = snapshot?.project.id === projectID;

  useEffect(() => {
    if (!projectID || !boardReady) return;
    const stream = new EventSource(api.getKanbanStreamURL(projectID, latestEventID));
    stream.addEventListener("kanban.event", () => {
      void loadBoard(projectID).catch((err) =>
        setError(String(err.message || err)),
      );
    });
    return () => stream.close();
  }, [projectID, boardReady, loadBoard]);

  const selectedTicket = useMemo(
    () => snapshot?.tickets.find((ticket) => ticket.id === selectedID) ?? null,
    [snapshot, selectedID],
  );
  const comments = useMemo(
    () =>
      selectedTicket
        ? (snapshot?.comments ?? []).filter(
            (comment) => comment.ticket_id === selectedTicket.id,
          )
        : [],
    [snapshot, selectedTicket],
  );
  const sourceRefs = useMemo(
    () =>
      selectedTicket
        ? (snapshot?.source_refs ?? []).filter(
            (ref) => ref.ticket_id === selectedTicket.id,
          )
        : [],
    [snapshot, selectedTicket],
  );
  const filteredTickets = useMemo(() => {
    const needle = query.trim().toLowerCase();
    const tickets = snapshot?.tickets ?? [];
    if (!needle) return tickets;
    return tickets.filter((ticket) =>
      [ticket.title, ticket.description, ticket.assignee, ticket.priority, ticket.id]
        .join(" ")
        .toLowerCase()
        .includes(needle),
    );
  }, [snapshot, query]);

  async function createProject() {
    const name = newProjectName.trim();
    if (!name) return;
    setError("");
    const response = await api.createKanbanProject({ name });
    await loadProjects();
    setProjectID(response.project.id);
    setNewProjectName("");
    setCreatingProject(false);
  }

  async function createTicket() {
    const title = newTicketTitle.trim();
    if (!title || !projectID) return;
    setError("");
    await api.createKanbanTicket({
      project_id: projectID,
      title,
      description: newTicketDescription.trim(),
    });
    setNewTicketTitle("");
    setNewTicketDescription("");
    await loadBoard(projectID);
  }

  async function moveTicket(ticket: KanbanTicket, status: KanbanTicketStatus) {
    if (ticket.status === status) return;
    if (!isStatusTransitionAllowed(ticket.status, status)) {
      setError(
        `Cannot move ticket from ${STATUS_LABEL[ticket.status]} to ${STATUS_LABEL[status]}`,
      );
      return;
    }
    setSnapshot((current) =>
      current
        ? {
            ...current,
            tickets: current.tickets.map((item) =>
              item.id === ticket.id ? { ...item, status } : item,
            ),
          }
        : current,
    );
    try {
      await api.updateKanbanTicket(ticket.id, { status });
      await loadBoard(projectID);
    } catch (err) {
      setError(String((err as Error).message || err));
      await loadBoard(projectID);
    }
  }

  if (loading) {
    return (
      <div className="flex h-[50vh] items-center justify-center text-muted-foreground">
        <Spinner />
      </div>
    );
  }

  return (
    <div className="flex h-[calc(100dvh-7rem)] min-h-0 flex-col gap-3">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex min-w-0 items-center gap-3">
          <Columns3 className="h-5 w-5 text-primary" />
          <div>
            <h1 className="text-lg font-semibold text-foreground">Kanban</h1>
            <p className="text-xs text-muted-foreground">
              {snapshot ? `${snapshot.project.name} · ${snapshot.tickets.length} tickets` : "Internal RSI tickets"}
            </p>
          </div>
        </div>
        <div className="flex min-w-0 flex-wrap items-center gap-2">
          <select
            className="h-9 border border-border bg-background px-2 text-sm text-foreground"
            value={projectID}
            onChange={(event) => setProjectID(event.target.value)}
          >
            {projects.map((project) => (
              <option key={project.id} value={project.id}>
                {project.name}
              </option>
            ))}
          </select>
          <Button size="sm" onClick={() => setCreatingProject(true)}>
            <Plus className="mr-1 h-4 w-4" />
            Project
          </Button>
        </div>
      </div>

      {error ? (
        <div className="border border-destructive/60 bg-destructive/10 px-3 py-2 text-sm text-destructive">
          {error}
        </div>
      ) : null}

      {creatingProject ? (
        <div className="flex flex-wrap items-center gap-2 border border-border bg-muted/20 p-3">
          <Input
            value={newProjectName}
            onChange={(event) => setNewProjectName(event.target.value)}
            placeholder="Project name"
            className="max-w-sm"
          />
          <Button size="sm" onClick={createProject}>Create</Button>
          <Button ghost size="sm" onClick={() => setCreatingProject(false)}>Cancel</Button>
        </div>
      ) : null}

      {projects.length === 0 ? (
        <div className="flex flex-1 items-center justify-center border border-border bg-muted/10 p-6 text-center">
          <div className="max-w-sm">
            <p className="mb-3 text-sm text-muted-foreground">Create the first RSI project to start tracking tickets.</p>
            <div className="flex gap-2">
              <Input value={newProjectName} onChange={(event) => setNewProjectName(event.target.value)} placeholder="Project name" />
              <Button onClick={createProject}>Create</Button>
            </div>
          </div>
        </div>
      ) : (
        <>
          <div className="flex flex-wrap items-start gap-2 border border-border bg-muted/10 p-3">
            <div className="relative min-w-[220px] flex-1">
              <Search className="pointer-events-none absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input
                value={query}
                onChange={(event) => setQuery(event.target.value)}
                placeholder="Search tickets"
                className="pl-8"
              />
            </div>
            <Input
              value={newTicketTitle}
              onChange={(event) => setNewTicketTitle(event.target.value)}
              placeholder="New ticket title"
              className="min-w-[220px] flex-1"
            />
            <Input
              value={newTicketDescription}
              onChange={(event) => setNewTicketDescription(event.target.value)}
              placeholder="Description"
              className="min-w-[220px] flex-1"
            />
            <Button onClick={createTicket} disabled={!newTicketTitle.trim()}>
              <Plus className="mr-1 h-4 w-4" />
              Ticket
            </Button>
          </div>

          <div className="min-h-0 flex-1 overflow-x-auto overflow-y-hidden">
            <div className="grid h-full min-w-[1120px] grid-cols-6 gap-3">
              {STATUSES.map((column) => {
                const Icon = column.icon;
                const tickets = filteredTickets.filter(
                  (ticket) => ticket.status === column.status,
                );
                return (
                  <section
                    key={column.status}
                    className="flex min-h-0 flex-col border border-border bg-muted/10"
                    onDragOver={(event) => event.preventDefault()}
                    onDrop={(event) => {
                      const id = event.dataTransfer.getData("text/plain");
                      const ticket = snapshot?.tickets.find((item) => item.id === id);
                      if (ticket) void moveTicket(ticket, column.status);
                    }}
                  >
                    <header className="flex items-center justify-between border-b border-border px-3 py-2">
                      <span className="flex items-center gap-2 text-sm font-medium text-foreground">
                        <Icon className="h-4 w-4" />
                        {column.label}
                      </span>
                      <Badge tone="outline">{tickets.length}</Badge>
                    </header>
                    <div className="min-h-0 flex-1 space-y-2 overflow-y-auto p-2">
                      {tickets.map((ticket) => (
                        <TicketCard
                          key={ticket.id}
                          ticket={ticket}
                          selected={ticket.id === selectedID}
                          onSelect={() => setSelectedID(ticket.id)}
                        />
                      ))}
                    </div>
                  </section>
                );
              })}
            </div>
          </div>
        </>
      )}

      {selectedTicket ? (
        <TicketDrawer
          ticket={selectedTicket}
          comments={comments}
          sourceRefs={sourceRefs}
          onClose={() => setSelectedID("")}
          onComment={async (body) => {
            await api.commentKanbanTicket(selectedTicket.id, { body });
            await loadBoard(projectID);
          }}
        />
      ) : null}
    </div>
  );
}

function TicketCard({
  ticket,
  selected,
  onSelect,
}: {
  ticket: KanbanTicket;
  selected: boolean;
  onSelect: () => void;
}) {
  return (
    <button
      draggable
      onDragStart={(event) => event.dataTransfer.setData("text/plain", ticket.id)}
      onClick={onSelect}
      className={cn(
        "w-full border bg-background/70 p-3 text-left transition-colors",
        selected ? "border-primary" : "border-border hover:border-foreground/40",
      )}
    >
      <div className="mb-2 line-clamp-3 text-sm font-medium text-foreground">
        {ticket.title}
      </div>
      {ticket.description ? (
        <p className="mb-3 line-clamp-3 text-xs text-muted-foreground">
          {ticket.description}
        </p>
      ) : null}
      <div className="flex flex-wrap items-center gap-1.5 text-[11px] text-muted-foreground">
        {ticket.priority ? <Badge tone="outline">{ticket.priority}</Badge> : null}
        {ticket.assignee ? <span>{ticket.assignee}</span> : null}
        <span>{timeAgo(new Date(ticket.updated_at).getTime() / 1000)}</span>
      </div>
    </button>
  );
}

function TicketDrawer({
  ticket,
  comments,
  sourceRefs,
  onClose,
  onComment,
}: {
  ticket: KanbanTicket;
  comments: KanbanTicketComment[];
  sourceRefs: KanbanTicketSourceRef[];
  onClose: () => void;
  onComment: (body: string) => Promise<void>;
}) {
  const [comment, setComment] = useState("");
  return (
    <div className="fixed inset-y-0 right-0 z-50 flex w-full max-w-xl flex-col border-l border-border bg-background shadow-xl">
      <header className="flex items-start justify-between gap-3 border-b border-border p-4">
        <div>
          <Badge tone="outline">{STATUS_LABEL[ticket.status]}</Badge>
          <h2 className="mt-3 text-base font-semibold text-foreground">{ticket.title}</h2>
          <p className="mt-1 text-xs text-muted-foreground">{ticket.id}</p>
        </div>
        <Button ghost size="icon" onClick={onClose}>
          <X className="h-4 w-4" />
        </Button>
      </header>
      <div className="min-h-0 flex-1 space-y-5 overflow-y-auto p-4">
        <section>
          <h3 className="mb-2 text-xs font-medium uppercase tracking-wide text-muted-foreground">Description</h3>
          <p className="whitespace-pre-wrap text-sm text-foreground">
            {ticket.description || "No description."}
          </p>
        </section>
        <section>
          <h3 className="mb-2 text-xs font-medium uppercase tracking-wide text-muted-foreground">Source refs</h3>
          <div className="space-y-2">
            {sourceRefs.length === 0 ? (
              <p className="text-sm text-muted-foreground">No source refs.</p>
            ) : (
              sourceRefs.map((ref) => (
                <div key={ref.id} className="border border-border p-2 text-xs text-muted-foreground">
                  <div className="flex items-center gap-2 text-foreground">
                    <MessageSquare className="h-3.5 w-3.5" />
                    {ref.source_type} · {ref.action_kind}
                  </div>
                  <div className="mt-1 break-all">
                    {ref.channel_id ? `channel ${ref.channel_id}` : ""}
                    {ref.thread_ts ? ` · thread ${ref.thread_ts}` : ""}
                    {ref.message_ts ? ` · message ${ref.message_ts}` : ""}
                  </div>
                </div>
              ))
            )}
          </div>
        </section>
        <section>
          <h3 className="mb-2 text-xs font-medium uppercase tracking-wide text-muted-foreground">Comments</h3>
          <div className="space-y-2">
            {comments.map((item) => (
              <div key={item.id} className="border border-border p-2 text-sm">
                <div className="mb-1 text-xs text-muted-foreground">
                  {item.actor_display || item.actor_id} · {timeAgo(new Date(item.created_at).getTime() / 1000)}
                </div>
                <p className="whitespace-pre-wrap text-foreground">{item.body}</p>
              </div>
            ))}
          </div>
          <div className="mt-3 flex gap-2">
            <Input value={comment} onChange={(event) => setComment(event.target.value)} placeholder="Add comment" />
            <Button
              onClick={async () => {
                const body = comment.trim();
                if (!body) return;
                setComment("");
                await onComment(body);
              }}
            >
              Send
            </Button>
          </div>
        </section>
      </div>
    </div>
  );
}
