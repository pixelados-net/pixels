--liquibase formatted sql

--changeset pixels:pixels-catalog-0003-create-catalog-item-limited-units
create table catalog_item_limited_units (
    id bigint generated always as identity primary key,
    catalog_item_id bigint not null references catalog_items(id),
    unit_number integer not null,
    owner_player_id bigint null references players(id),
    furniture_item_id bigint null references furniture_items(id),
    sold_at timestamptz null,
    constraint catalog_item_limited_units_number_positive_chk check (unit_number > 0),
    constraint catalog_item_limited_units_sale_state_chk check (
        (owner_player_id is null and furniture_item_id is null and sold_at is null)
        or
        (owner_player_id is not null and sold_at is not null)
    ),
    unique (catalog_item_id, unit_number)
);

create index catalog_item_limited_units_available_idx
on catalog_item_limited_units (catalog_item_id, unit_number)
where owner_player_id is null;
--rollback drop table if exists catalog_item_limited_units;
