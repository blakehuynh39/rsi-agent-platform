create index if not exists conversation_updated_id_idx
  on conversation (updated_at desc, id asc);

create index if not exists trace_summary_hermes_last_active_idx
  on trace_summary (
    (case
      when ended_at is null or ended_at <= timestamptz '0001-01-02 00:00:00+00'
      then started_at
      else ended_at
    end) desc,
    trace_id asc
  );

create index if not exists execution_ledger_event_recent_trace_idx
  on execution_ledger_event (recorded_at desc, trace_id, execution_id, seq, id)
  where trace_id <> '';

create index if not exists execution_ledger_event_latest_by_trace_idx
  on execution_ledger_event (trace_id asc, recorded_at desc, execution_id desc, seq desc, id desc)
  where trace_id <> '';
