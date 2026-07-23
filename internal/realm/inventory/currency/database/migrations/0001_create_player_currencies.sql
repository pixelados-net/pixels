--liquibase formatted sql

--changeset pixels:pixels-inventory-currency-0001-create-player-currencies
create table player_currencies (
    player_id bigint not null references players(id),
    currency_type integer not null,
    amount bigint not null default 0,
    updated_at timestamptz not null default now(),
    version bigint not null default 1,
    primary key (player_id, currency_type),
    constraint player_currencies_amount_non_negative_chk check (amount >= 0),
    constraint player_currencies_version_positive_chk check (version > 0)
);
--rollback drop table if exists player_currencies;
