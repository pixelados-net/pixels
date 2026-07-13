--liquibase formatted sql

--changeset pixels:pixels-trade-0001-create-trade-audit
create table trade_audit_logs (
    id bigint generated always as identity primary key,
    room_id bigint not null references rooms(id),
    first_player_id bigint not null references players(id),
    second_player_id bigint not null references players(id),
    first_ip inet null,
    second_ip inet null,
    first_item_ids bigint[] not null default '{}',
    second_item_ids bigint[] not null default '{}',
    first_redeemable_credits bigint not null default 0,
    second_redeemable_credits bigint not null default 0,
    created_at timestamptz not null default now()
);
create index trade_audit_logs_first_player_idx on trade_audit_logs(first_player_id,created_at desc);
create index trade_audit_logs_second_player_idx on trade_audit_logs(second_player_id,created_at desc);
--rollback drop table if exists trade_audit_logs;
