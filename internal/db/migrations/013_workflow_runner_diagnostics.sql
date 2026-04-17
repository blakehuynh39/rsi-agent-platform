alter table workflow
  add column if not exists runner_diagnostics jsonb not null default '{}'::jsonb;
