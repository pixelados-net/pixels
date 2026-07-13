--liquibase formatted sql

--changeset pixels:pixels-marketplace-0001-create-marketplace
create table marketplace_tokens (
    player_id bigint primary key references players(id),
    amount integer not null default 0 check (amount >= 0),
    updated_at timestamptz not null default now()
);

create table marketplace_listings (
    id bigint generated always as identity primary key,
    seller_player_id bigint not null references players(id),
    buyer_player_id bigint null references players(id),
    furniture_item_id bigint not null references furniture_items(id),
    furniture_definition_id bigint not null references furniture_definitions(id),
    raw_price bigint not null check (raw_price > 0),
    state smallint not null default 0 check (state between 0 and 2),
    expires_at timestamptz not null,
    sold_at timestamptz null,
    closed_at timestamptz null,
    redeemed_at timestamptz null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    version bigint not null default 1 check (version > 0)
);

create unique index marketplace_listings_open_item_uidx on marketplace_listings (furniture_item_id) where state = 0;
create index marketplace_listings_search_idx on marketplace_listings (furniture_definition_id, raw_price, expires_at) where state = 0;
create index marketplace_listings_seller_idx on marketplace_listings (seller_player_id, state, created_at desc);
create index marketplace_listings_expiry_idx on marketplace_listings (expires_at) where state = 0;

create table marketplace_daily_stats (
    furniture_definition_id bigint not null references furniture_definitions(id),
    day date not null,
    average_raw_price bigint not null check (average_raw_price >= 0),
    sold_count integer not null check (sold_count >= 0),
    primary key (furniture_definition_id, day)
);
--rollback drop table if exists marketplace_daily_stats; drop table if exists marketplace_listings; drop table if exists marketplace_tokens;
