--liquibase formatted sql

--changeset pixels:pixels-catalog-0012-room-bundles
alter table catalog_items alter column definition_id drop not null;
alter table catalog_items add column room_bundle_template_room_id bigint null references rooms(id) on delete restrict;
alter table catalog_items add constraint catalog_items_product_kind_chk check (
    (room_bundle_template_room_id is null and definition_id is not null)
    or
    (room_bundle_template_room_id is not null and definition_id is null and amount = 0 and limited_stack = 0 and limited_sells = 0 and not giftable and not bundle_discount_enabled)
);
create index catalog_items_room_bundle_template_idx on catalog_items (room_bundle_template_room_id) where room_bundle_template_room_id is not null and deleted_at is null;

create table room_bundle_purchases (
    id bigint generated always as identity primary key,
    catalog_item_id bigint not null references catalog_items(id) on delete restrict,
    template_room_id bigint not null references rooms(id) on delete restrict,
    created_room_id bigint not null references rooms(id) on delete restrict,
    buyer_player_id bigint not null references players(id) on delete restrict,
    furniture_count integer not null default 0,
    purchased_at timestamptz not null default now(),
    constraint room_bundle_purchases_furniture_count_chk check (furniture_count >= 0),
    unique (created_room_id)
);
create index room_bundle_purchases_buyer_idx on room_bundle_purchases (buyer_player_id, purchased_at desc);
--rollback drop table if exists room_bundle_purchases; drop index if exists catalog_items_room_bundle_template_idx; alter table catalog_items drop constraint if exists catalog_items_product_kind_chk; alter table catalog_items drop column if exists room_bundle_template_room_id; alter table catalog_items alter column definition_id set not null;
