--liquibase formatted sql

--changeset pixels:pixels-subscription-0001-create-memberships
create table subscription_memberships (
    player_id bigint primary key references players(id),
    level smallint not null default 0,
    started_at timestamptz null,
    expires_at timestamptz null,
    last_payday_at timestamptz null,
    lifetime_active_seconds bigint not null default 0,
    gifts_earned integer not null default 0,
    gifts_claimed integer not null default 0,
    version bigint not null default 1,
    constraint subscription_memberships_level_chk check (level between 0 and 2)
);
create table subscription_club_offers (
    id bigint generated always as identity primary key,
    name text not null unique,
    day_count integer not null,
    price_credits bigint not null default 0,
    price_points bigint not null default 0,
    points_type integer not null default -1,
    is_vip boolean not null default false,
    is_deal boolean not null default false,
    enabled boolean not null default true,
    order_num integer not null default 0,
    constraint subscription_club_offers_days_chk check (day_count > 0)
);
--rollback drop table if exists subscription_club_offers; drop table if exists subscription_memberships;
