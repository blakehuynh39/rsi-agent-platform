update effect_execution
set queue_name = 'action'
where machine_kind = 'action'
  and queue_name <> 'action';
