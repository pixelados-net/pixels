--liquibase formatted sql

--changeset pixels:pixels-inventory-currency-0002-create-currency-ledger-entries
create table currency_ledger_entries (
    id bigint generated always as identity primary key,
    player_id bigint not null references players(id),
    currency_type integer not null,
    delta bigint not null,
    balance_after bigint not null,
    reason text not null default '',
    actor_kind text not null,
    actor_id bigint null,
    created_at timestamptz not null default now(),
    constraint currency_ledger_entries_actor_kind_chk check (actor_kind in ('system', 'admin', 'player'))
);

create index currency_ledger_entries_player_created_idx
on currency_ledger_entries (player_id, created_at desc);
--rollback drop table if exists currency_ledger_entries;
