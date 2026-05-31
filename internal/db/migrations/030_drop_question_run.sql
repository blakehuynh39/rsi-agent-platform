delete from effect_execution where machine_kind = 'question_run';

delete from command_receipt where machine_kind = 'question_run';

delete from domain_event where machine_kind = 'question_run';

drop index if exists question_run_trace_idx;
drop index if exists question_run_workflow_idx;
drop table if exists question_run;
