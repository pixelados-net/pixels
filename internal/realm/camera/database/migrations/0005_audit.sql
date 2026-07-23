--liquibase formatted sql

--changeset pixels:pixels-camera-0005-audit
create table camera_audit_log (
    id bigserial primary key,
    actor_player_id bigint not null references players(id),
    action text not null,
    entity_id bigint null,
    reason text not null,
    created_at timestamptz not null default now()
);
create index camera_audit_actor_idx on camera_audit_log(actor_player_id, created_at desc);
--rollback drop table if exists camera_audit_log;
