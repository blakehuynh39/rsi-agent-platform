create index if not exists trace_summary_conversation_started_idx on trace_summary (conversation_id, started_at desc);
create index if not exists conversation_entry_conv_created_id_idx on conversation_entry (conversation_id, created_at asc, id asc);
