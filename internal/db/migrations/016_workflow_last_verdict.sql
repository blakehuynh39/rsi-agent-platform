alter table workflow
  add column if not exists last_verdict text not null default '';
