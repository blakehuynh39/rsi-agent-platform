alter table if exists source_mirror_record
  add column if not exists honcho_object_type text not null default '';

alter table if exists source_mirror_record
  add column if not exists honcho_object_id text not null default '';

update source_mirror_record
set honcho_object_type = 'message',
    honcho_object_id = honcho_message_id
where honcho_message_id <> ''
  and honcho_object_id = '';

create index if not exists source_mirror_record_honcho_object_idx
  on source_mirror_record (honcho_object_type, honcho_object_id)
  where honcho_object_id <> '';
