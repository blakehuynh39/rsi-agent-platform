create index if not exists execution_ledger_event_trace_page_idx
  on execution_ledger_event (trace_id, recorded_at desc, execution_id desc, seq desc, id desc);

create index if not exists trace_event_trace_started_idx
  on trace_event (trace_id, started_at asc);

create index if not exists artifact_trace_idx
  on artifact (trace_id, id);
