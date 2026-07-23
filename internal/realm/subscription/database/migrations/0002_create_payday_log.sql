--liquibase formatted sql

--changeset pixels:pixels-subscription-0002-create-payday-log
create table subscription_payday_log (
    id bigint generated always as identity primary key,
    player_id bigint not null references players(id),
    occurred_at timestamptz not null default now(),
    streak_days integer not null,
    credits_spent bigint not null,
    streak_bonus bigint not null,
    monthly_bonus bigint not null,
    total_awarded bigint not null,
    currency_type integer not null default -1,
    claimed boolean not null default false
);
create index subscription_payday_log_unclaimed_idx on subscription_payday_log (player_id) where claimed = false;
--rollback drop table if exists subscription_payday_log;
