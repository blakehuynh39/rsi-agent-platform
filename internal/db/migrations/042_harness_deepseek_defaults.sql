-- Move uncustomized legacy harness profiles to the direct DeepSeek default.
-- Profiles whose model was edited away from the old seeded default are left
-- untouched so operator customizations remain authoritative.

update harness_profile
set model = 'deepseek/deepseek-v4-pro',
    reasoning_effort = 'xhigh',
    updated_at = now()
where id in (
  'harness-profile-prod',
  'harness-profile-proactive',
  'harness-profile-eval',
  'harness-profile-proposal'
)
and model = 'openai/gpt-5.4';
