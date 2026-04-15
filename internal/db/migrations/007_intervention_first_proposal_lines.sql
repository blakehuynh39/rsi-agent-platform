alter table if exists proposal
  add column if not exists recommended_intervention_kind text not null default 'repo_change',
  add column if not exists recommended_intervention_rationale text not null default '',
  add column if not exists target_surface text not null default '',
  add column if not exists touched_files jsonb not null default '[]'::jsonb,
  add column if not exists validation_plan text not null default '',
  add column if not exists material_risk_summary text not null default '',
  add column if not exists recommended_disposition text not null default '';

alter table if exists proposal_review
  add column if not exists scope text not null default 'line';

update proposal
set recommended_intervention_kind = case
      when coalesce(nullif(target_layer, ''), 'repo_change') = 'harness_overlay' then 'harness_overlay'
      else 'repo_change'
    end,
    recommended_intervention_rationale = case
      when coalesce(nullif(summary, ''), '') <> '' then summary
      else 'Intervention recommendation inferred from candidate evidence.'
    end,
    target_surface = case
      when coalesce(nullif(proposed_scope, ''), '') <> '' then proposed_scope
      when coalesce(nullif(target_kind, ''), '') <> '' and coalesce(nullif(target_ref, ''), '') <> '' then target_kind || ':' || target_ref
      when coalesce(nullif(target_ref, ''), '') <> '' then target_ref
      when coalesce(nullif(target_kind, ''), '') <> '' then target_kind
      else 'unspecified_target_surface'
    end,
    validation_plan = case
      when coalesce(nullif(target_layer, ''), 'repo_change') = 'harness_overlay' then 'Generate a bounded runtime overlay, validate behavior in the target role, and activate only if the change remains inside the approved scope.'
      else 'Generate a bounded repo change, validate it in sandbox, and only then open a draft PR.'
    end,
    material_risk_summary = case
      when coalesce(nullif(risk_tier, ''), '') <> '' and coalesce(nullif(target_layer, ''), 'repo_change') = 'harness_overlay' then risk_tier || ' risk intervention on runtime harness surface.'
      when coalesce(nullif(risk_tier, ''), '') <> '' then risk_tier || ' risk repo-change intervention.'
      when coalesce(nullif(target_layer, ''), 'repo_change') = 'harness_overlay' then 'medium risk intervention on runtime harness surface.'
      else 'medium risk repo-change intervention.'
    end,
    recommended_disposition = 'approve_intervention'
where coalesce(recommended_intervention_kind, '') = ''
   or coalesce(target_surface, '') = ''
   or coalesce(validation_plan, '') = ''
   or coalesce(material_risk_summary, '') = ''
   or coalesce(recommended_disposition, '') = '';
