alter table if exists improvement_candidate add column if not exists target_layer text not null default 'repo_change';
alter table if exists improvement_candidate add column if not exists target_kind text not null default 'repo';
alter table if exists improvement_candidate add column if not exists target_ref text not null default '';

alter table if exists proposal add column if not exists target_layer text not null default 'repo_change';
alter table if exists proposal add column if not exists target_kind text not null default 'repo';
alter table if exists proposal add column if not exists target_ref text not null default '';

create table if not exists harness_profile (
  id text primary key,
  role text not null,
  name text not null,
  description text not null default '',
  model text not null default '',
  reasoning_effort text not null default '',
  prompt_fragments jsonb not null default '[]'::jsonb,
  few_shot_snippets jsonb not null default '[]'::jsonb,
  tool_preference_order jsonb not null default '[]'::jsonb,
  retrieval_bias text not null default '',
  reasoning_verbosity text not null default '',
  memory_read_enabled boolean not null default true,
  memory_write_enabled boolean not null default true,
  repo_ref text not null default '',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index if not exists harness_profile_role_idx on harness_profile (role);

create table if not exists harness_overlay (
  id text primary key,
  profile_id text not null references harness_profile(id),
  role text not null,
  version text not null,
  status text not null,
  target_kind text not null default '',
  target_ref text not null default '',
  proposal_id text,
  prompt_fragments jsonb not null default '[]'::jsonb,
  few_shot_snippets jsonb not null default '[]'::jsonb,
  tool_preference_order jsonb not null default '[]'::jsonb,
  retrieval_bias text not null default '',
  reasoning_verbosity text not null default '',
  memory_read_enabled boolean,
  memory_write_enabled boolean,
  created_by text not null default '',
  approved_by text not null default '',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  activated_at timestamptz
);

create unique index if not exists harness_overlay_role_version_idx on harness_overlay (role, version);
create index if not exists harness_overlay_role_status_idx on harness_overlay (role, status, updated_at desc);

create table if not exists harness_experiment (
  id text primary key,
  profile_id text not null references harness_profile(id),
  overlay_id text references harness_overlay(id),
  proposal_id text,
  role text not null,
  status text not null,
  summary text not null default '',
  metrics jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index if not exists harness_experiment_role_idx on harness_experiment (role, updated_at desc);

create table if not exists harness_session_binding (
  role text not null,
  scope_kind text not null,
  scope_id text not null,
  parent_scope_kind text not null default '',
  parent_scope_id text not null default '',
  hermes_session_id text not null,
  parent_session_id text not null default '',
  memory_backend text not null default '',
  assistant_peer_id text not null default '',
  user_peer_id text not null default '',
  harness_profile_id text not null default '',
  effective_overlay_id text not null default '',
  effective_overlay_version text not null default '',
  last_used_at timestamptz not null default now(),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  primary key (role, scope_kind, scope_id)
);

create index if not exists harness_session_binding_last_used_idx on harness_session_binding (last_used_at desc);

create table if not exists harness_execution (
  id text primary key,
  trace_id text,
  proposal_id text,
  role text not null,
  session_scope_kind text not null,
  session_scope_id text not null,
  hermes_session_id text not null,
  parent_session_id text not null default '',
  harness_profile_id text not null default '',
  effective_overlay_id text not null default '',
  effective_overlay_version text not null default '',
  memory_backend text not null default '',
  memory_reads jsonb not null default '[]'::jsonb,
  memory_writes jsonb not null default '[]'::jsonb,
  created_at timestamptz not null default now()
);

create index if not exists harness_execution_trace_idx on harness_execution (trace_id, created_at desc);
create index if not exists harness_execution_proposal_idx on harness_execution (proposal_id, created_at desc);
create index if not exists harness_execution_role_scope_idx on harness_execution (role, session_scope_kind, session_scope_id, created_at desc);

insert into harness_profile (
  id,
  role,
  name,
  description,
  model,
  reasoning_effort,
  prompt_fragments,
  few_shot_snippets,
  tool_preference_order,
  retrieval_bias,
  reasoning_verbosity,
  memory_read_enabled,
  memory_write_enabled,
  repo_ref,
  created_at,
  updated_at
)
values
  (
    'harness-profile-prod',
    'prod',
    'Production Operator',
    'Live conversation and incident workflow agent with durable memory and explicit evidence-first reasoning.',
    'openai/gpt-5.4',
    'xhigh',
    '["Ground answers in explicit evidence. Prefer concrete repo, Slack, and tool context over generic advice."]'::jsonb,
    '[]'::jsonb,
    '["repo.context","knowledge.context","github.repo_activity","sentry.lookup","kubernetes.logs"]'::jsonb,
    'canonical_then_working_then_session',
    'verbose',
    true,
    true,
    'main',
    now(),
    now()
  ),
  (
    'harness-profile-proactive',
    'proactive',
    'Proactive Thread Agent',
    'Monitors and joins conversations when evidence justifies intervention.',
    'openai/gpt-5.4',
    'xhigh',
    '["Intervene only when the evidence supports a useful reply or workflow launch."]'::jsonb,
    '[]'::jsonb,
    '["knowledge.context","repo.context","github.repo_activity"]'::jsonb,
    'canonical_then_session',
    'verbose',
    true,
    true,
    'main',
    now(),
    now()
  ),
  (
    'harness-profile-eval',
    'eval',
    'Eval Analyst',
    'Summarizes failures, compares traces, and improves recurring eval lines without hiding uncertainty.',
    'openai/gpt-5.4',
    'xhigh',
    '["Focus on observable evidence, failure patterns, and novelty relative to prior rejected proposals."]'::jsonb,
    '[]'::jsonb,
    '["knowledge.context"]'::jsonb,
    'canonical_then_session',
    'verbose',
    true,
    true,
    'main',
    now(),
    now()
  ),
  (
    'harness-profile-proposal',
    'proposal',
    'Proposal Materializer',
    'Turns approved candidate lines into governed repo-change or overlay-ready reasoning with prior memory context.',
    'openai/gpt-5.4',
    'xhigh',
    '["Respect proposal memory, review rationale, and rollback expectations before materializing work."]'::jsonb,
    '[]'::jsonb,
    '["knowledge.context","repo.context"]'::jsonb,
    'canonical_then_working_then_session',
    'verbose',
    true,
    true,
    'main',
    now(),
    now()
  )
on conflict (id) do update set
  role = excluded.role,
  name = excluded.name,
  description = excluded.description,
  model = excluded.model,
  reasoning_effort = excluded.reasoning_effort,
  prompt_fragments = excluded.prompt_fragments,
  few_shot_snippets = excluded.few_shot_snippets,
  tool_preference_order = excluded.tool_preference_order,
  retrieval_bias = excluded.retrieval_bias,
  reasoning_verbosity = excluded.reasoning_verbosity,
  memory_read_enabled = excluded.memory_read_enabled,
  memory_write_enabled = excluded.memory_write_enabled,
  repo_ref = excluded.repo_ref,
  updated_at = excluded.updated_at;
