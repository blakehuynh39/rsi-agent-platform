alter table external_tool_action
  add column if not exists result_payload jsonb not null default '{}'::jsonb;
