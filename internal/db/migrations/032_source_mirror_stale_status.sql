alter table if exists source_mirror_record
  drop constraint if exists source_mirror_record_status_check;

alter table if exists source_mirror_record
  add constraint source_mirror_record_status_check
  check (status in ('pending', 'complete', 'failed', 'stale'));
