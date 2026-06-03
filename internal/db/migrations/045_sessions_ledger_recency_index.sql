create index if not exists execution_ledger_event_session_recency_idx
  on execution_ledger_event (recorded_at desc, trace_id asc, execution_id desc, seq desc, id desc)
  where trace_id <> '';
