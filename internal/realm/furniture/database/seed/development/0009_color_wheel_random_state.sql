--liquibase formatted sql

--changeset pixels:pixels-furniture-seed-development-0009-color-wheel-random-state context:development
update furniture_definitions
set interaction_type = 'random_state',
    custom_params = 'states=10,delay=1200',
    metadata = metadata || '{"test_variant":true,"behavior":"random_state"}'::jsonb,
    updated_at = now()
where id = 27 and name = 'wf_colorwheel';
--rollback update furniture_definitions set interaction_type = 'default', custom_params = '', metadata = metadata - 'behavior', updated_at = now() where id = 27;
