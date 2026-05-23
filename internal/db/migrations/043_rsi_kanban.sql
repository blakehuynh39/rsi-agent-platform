create table if not exists kanban_project (
  id text primary key,
  slug text not null unique,
  name text not null,
  description text not null default '',
  state text not null default 'active',
  metadata jsonb not null default '{}'::jsonb,
  created_at timestamptz not null,
  updated_at timestamptz not null,
  constraint ck_kanban_project_state check (state in ('active', 'archived'))
);

create table if not exists kanban_board (
  id text primary key,
  project_id text not null references kanban_project(id) on delete cascade,
  slug text not null,
  name text not null,
  is_default boolean not null default false,
  metadata jsonb not null default '{}'::jsonb,
  created_at timestamptz not null,
  updated_at timestamptz not null,
  unique (id, project_id),
  unique (project_id, slug)
);

create unique index if not exists kanban_board_one_default_per_project_idx
  on kanban_board (project_id)
  where is_default;

create table if not exists kanban_ticket (
  id text primary key,
  project_id text not null,
  board_id text not null,
  title text not null,
  description text not null default '',
  status text not null,
  priority text not null default '',
  assignee text not null default '',
  created_by text not null,
  metadata jsonb not null default '{}'::jsonb,
  created_at timestamptz not null,
  updated_at timestamptz not null,
  completed_at timestamptz,
  archived_at timestamptz,
  unique (id, project_id),
  constraint fk_kanban_ticket_board_project
    foreign key (board_id, project_id) references kanban_board(id, project_id) on delete cascade,
  constraint ck_kanban_ticket_status check (status in ('triage', 'todo', 'in_progress', 'blocked', 'done', 'archived'))
);

create index if not exists kanban_ticket_project_status_idx
  on kanban_ticket (project_id, status, updated_at desc);

create table if not exists kanban_ticket_comment (
  id text primary key,
  project_id text not null,
  ticket_id text not null,
  body text not null,
  actor_type text not null,
  actor_id text not null,
  actor_display text not null default '',
  source_surface text not null default '',
  metadata jsonb not null default '{}'::jsonb,
  created_at timestamptz not null,
  constraint fk_kanban_comment_ticket_project
    foreign key (ticket_id, project_id) references kanban_ticket(id, project_id) on delete cascade
);

create index if not exists kanban_ticket_comment_ticket_idx
  on kanban_ticket_comment (ticket_id, created_at asc);

create table if not exists kanban_ticket_link (
  id text primary key,
  project_id text not null,
  from_ticket_id text not null,
  to_ticket_id text not null,
  link_type text not null,
  created_by text not null,
  metadata jsonb not null default '{}'::jsonb,
  created_at timestamptz not null,
  constraint fk_kanban_link_from_ticket_project
    foreign key (from_ticket_id, project_id) references kanban_ticket(id, project_id) on delete cascade,
  constraint fk_kanban_link_to_ticket_project
    foreign key (to_ticket_id, project_id) references kanban_ticket(id, project_id) on delete cascade,
  constraint ck_kanban_link_not_self check (from_ticket_id <> to_ticket_id)
);

create unique index if not exists kanban_ticket_link_unique_idx
  on kanban_ticket_link (project_id, from_ticket_id, to_ticket_id, link_type);

create table if not exists kanban_ticket_source_ref (
  id text primary key,
  project_id text not null,
  ticket_id text not null,
  source_type text not null,
  action_kind text not null,
  team_id text not null default '',
  channel_id text not null default '',
  thread_ts text not null default '',
  message_ts text not null default '',
  permalink text not null default '',
  conversation_id text not null default '',
  trace_id text not null default '',
  workflow_id text not null default '',
  proposal_id text not null default '',
  metadata jsonb not null default '{}'::jsonb,
  created_at timestamptz not null,
  constraint fk_kanban_source_ref_ticket_project
    foreign key (ticket_id, project_id) references kanban_ticket(id, project_id) on delete cascade
);

create unique index if not exists kanban_ticket_source_ref_slack_idempotency_idx
  on kanban_ticket_source_ref (team_id, channel_id, thread_ts, message_ts, action_kind)
  where source_type = 'slack'
    and team_id <> ''
    and channel_id <> ''
    and thread_ts <> ''
    and message_ts <> ''
    and action_kind <> '';

create unique index if not exists kanban_ticket_source_ref_slack_idempotency_no_team_idx
  on kanban_ticket_source_ref (channel_id, thread_ts, message_ts, action_kind)
  where source_type = 'slack'
    and team_id = ''
    and channel_id <> ''
    and thread_ts <> ''
    and message_ts <> ''
    and action_kind <> '';

create index if not exists kanban_ticket_source_ref_ticket_idx
  on kanban_ticket_source_ref (ticket_id, created_at asc);

create table if not exists kanban_ticket_event (
  id text primary key,
  project_id text not null,
  ticket_id text,
  event_type text not null,
  actor_type text not null,
  actor_id text not null,
  actor_display text not null default '',
  source_surface text not null default '',
  payload jsonb not null default '{}'::jsonb,
  created_at timestamptz not null,
  constraint fk_kanban_event_ticket_project
    foreign key (ticket_id, project_id) references kanban_ticket(id, project_id) on delete cascade
);

create index if not exists kanban_ticket_event_project_idx
  on kanban_ticket_event (project_id, created_at asc, id asc);

create table if not exists kanban_project_slack_route (
  id text primary key,
  project_id text not null references kanban_project(id) on delete cascade,
  team_id text not null default '',
  channel_id text not null,
  thread_ts text not null default '',
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create unique index if not exists kanban_project_slack_route_unique_idx
  on kanban_project_slack_route (team_id, channel_id, thread_ts);
