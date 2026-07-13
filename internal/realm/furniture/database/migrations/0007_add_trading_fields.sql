--liquibase formatted sql

--changeset pixels:pixels-furniture-0007-add-trading-fields
alter table furniture_definitions
    add column allow_trade boolean not null default true,
    add column allow_marketplace_sale boolean not null default true,
    add column redeemable_credits integer not null default 0,
    add constraint furniture_definitions_redeemable_credits_chk check (redeemable_credits >= 0);

alter table furniture_items
    add column limited_edition_number integer null,
    add column marketplace_reserved boolean not null default false,
    add constraint furniture_items_limited_edition_number_chk check (limited_edition_number is null or limited_edition_number > 0),
    add constraint furniture_items_marketplace_reservation_chk check (not marketplace_reserved or room_id is null);

create index furniture_items_marketplace_reserved_idx
on furniture_items (owner_player_id, marketplace_reserved)
where deleted_at is null and marketplace_reserved;
--rollback drop index if exists furniture_items_marketplace_reserved_idx; alter table furniture_items drop constraint if exists furniture_items_marketplace_reservation_chk, drop constraint if exists furniture_items_limited_edition_number_chk, drop column if exists marketplace_reserved, drop column if exists limited_edition_number; alter table furniture_definitions drop constraint if exists furniture_definitions_redeemable_credits_chk, drop column if exists redeemable_credits, drop column if exists allow_marketplace_sale, drop column if exists allow_trade;
