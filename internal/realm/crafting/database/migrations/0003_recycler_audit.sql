--liquibase formatted sql
--changeset pixels:pixels-crafting-0003-recycler-audit
create table crafting_recycler_prizes (
    tier smallint not null check(tier between 1 and 5),
    reward_definition_id bigint not null references furniture_definitions(id), primary key(tier,reward_definition_id)
);
create table crafting_audit (
    id bigint generated always as identity primary key,
    actor_player_id bigint null references players(id) on delete set null,
    action text not null, entity_kind text not null, entity_id bigint null, reason text not null,
    created_at timestamptz not null default now(), constraint crafting_audit_reason_chk check(char_length(reason) between 1 and 500)
);
create index crafting_audit_created_idx on crafting_audit(created_at desc,id desc);
--rollback drop table if exists crafting_audit; drop table if exists crafting_recycler_prizes;
