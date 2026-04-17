with duplicate_external_event as (
  select id
  from (
    select
      id,
      row_number() over (
        partition by conversation_id, event_id, entry_type
        order by created_at asc, id asc
      ) as row_num
    from conversation_entry
    where entry_type = 'external_event'
      and event_id is not null
  ) ranked
  where row_num > 1
),
duplicate_slack_action as (
  select id
  from (
    select
      id,
      row_number() over (
        partition by conversation_id, source_event_id, entry_type
        order by created_at asc, id asc
      ) as row_num
    from conversation_entry
    where entry_type = 'slack_action'
      and source_event_id is not null
  ) ranked
  where row_num > 1
)
delete from conversation_entry
where id in (
  select id from duplicate_external_event
  union all
  select id from duplicate_slack_action
);

create unique index if not exists conversation_entry_external_event_idx
  on conversation_entry (conversation_id, event_id, entry_type)
  where entry_type = 'external_event' and event_id is not null;

create unique index if not exists conversation_entry_slack_action_idx
  on conversation_entry (conversation_id, source_event_id, entry_type)
  where entry_type = 'slack_action' and source_event_id is not null;
