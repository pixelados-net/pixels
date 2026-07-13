--liquibase formatted sql

--changeset pixels:pixels-room-0008-room-bundle-templates
alter table rooms add column is_bundle_template boolean not null default false;
create index rooms_bundle_templates_idx on rooms (id) where is_bundle_template and deleted_at is null;
--rollback drop index if exists rooms_bundle_templates_idx; alter table rooms drop column if exists is_bundle_template;
