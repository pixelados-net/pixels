--liquibase formatted sql
--changeset pixels:pixels-progression-0006-audit
create table progression_audit (
    id bigint generated always as identity primary key, actor_player_id bigint null references players(id) on delete set null,
    action text not null, entity text not null, reason text not null, created_at timestamptz not null default now(),
    constraint progression_audit_reason_chk check(char_length(reason) between 1 and 500)
);
create index progression_audit_created_idx on progression_audit(created_at desc,id desc);
--rollback drop table if exists progression_audit;
