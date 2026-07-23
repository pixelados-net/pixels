--liquibase formatted sql

--changeset pixels:pixels-subscription-0005-create-calendar
create table calendar_campaigns (
    id bigint generated always as identity primary key,
    name text not null unique,
    image text not null default '',
    start_date date not null,
    day_count integer not null,
    enabled boolean not null default true,
    constraint calendar_campaigns_day_count_chk check (day_count between 1 and 31)
);
create table calendar_campaign_days (
    campaign_id bigint not null references calendar_campaigns(id) on delete cascade,
    day_number integer not null,
    product_definition_id bigint null references furniture_definitions(id),
    custom_image text not null default '',
    credits_reward bigint not null default 0,
    points_reward bigint not null default 0,
    points_type integer not null default -1,
    primary key (campaign_id, day_number),
    constraint calendar_campaign_days_day_number_chk check (day_number >= 0)
);
create table calendar_door_claims (
    campaign_id bigint not null references calendar_campaigns(id),
    player_id bigint not null references players(id),
    day_number integer not null,
    claimed_at timestamptz not null default now(),
    primary key (campaign_id, player_id, day_number)
);
create table calendar_seasonal_offers (
    offer_date date primary key,
    catalog_page_id bigint not null references catalog_pages(id),
    catalog_item_id bigint not null references catalog_items(id)
);
--rollback drop table if exists calendar_seasonal_offers; drop table if exists calendar_door_claims; drop table if exists calendar_campaign_days; drop table if exists calendar_campaigns;
