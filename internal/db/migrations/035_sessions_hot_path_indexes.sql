create index if not exists trace_summary_started_idx
  on trace_summary (started_at desc, trace_id asc);

create index if not exists reasoning_step_type_trace_created_idx
  on reasoning_step (lower(step_type), trace_id, created_at desc, id desc);

create index if not exists conversation_entry_conv_latest_nonempty_idx
  on conversation_entry (conversation_id, created_at desc, id desc)
  where btrim(body) <> '';
