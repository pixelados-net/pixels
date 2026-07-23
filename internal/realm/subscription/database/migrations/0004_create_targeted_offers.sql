--liquibase formatted sql

--changeset pixels:pixels-subscription-0004-create-targeted-offers
create table subscription_targeted_offers (
    id bigint generated always as identity primary key,
    catalog_item_id bigint not null references catalog_items(id),
    price_credits bigint not null default 0,
    price_points bigint not null default 0,
    points_type integer not null default -1,
    purchase_limit integer not null default 1,
    title_key text not null,
    description_key text not null,
    image_url text not null default '',
    icon_url text not null default '',
    enabled boolean not null default true,
    expires_at timestamptz null,
    order_num integer not null default 0,
    constraint subscription_targeted_offers_limit_chk check (purchase_limit > 0)
);
create table subscription_targeted_offer_progress (
    player_id bigint not null references players(id),
    offer_id bigint not null references subscription_targeted_offers(id),
    purchases_count integer not null default 0,
    last_viewed_at timestamptz null,
    dismissed boolean not null default false,
    primary key (player_id, offer_id)
);
--rollback drop table if exists subscription_targeted_offer_progress; drop table if exists subscription_targeted_offers;
