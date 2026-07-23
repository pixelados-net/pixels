--liquibase formatted sql

--changeset pixels:pixels-subscription-0003-create-gift-claims
create table subscription_club_gift_claims (
    player_id bigint not null references players(id),
    period_start timestamptz not null,
    claimed_item_id bigint not null references catalog_items(id),
    claimed_at timestamptz not null default now(),
    primary key (player_id, period_start)
);
--rollback drop table if exists subscription_club_gift_claims;
