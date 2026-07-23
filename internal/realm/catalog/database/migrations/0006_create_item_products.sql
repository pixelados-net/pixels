--liquibase formatted sql

--changeset pixels:pixels-catalog-0006-create-item-products
create table catalog_item_products (
    id bigint generated always as identity primary key,
    catalog_item_id bigint not null references catalog_items(id) on delete cascade,
    definition_id bigint not null references furniture_definitions(id),
    quantity integer not null default 1,
    order_num integer not null default 0,
    constraint catalog_item_products_quantity_chk check (quantity > 0),
    unique (catalog_item_id, definition_id)
);
create index catalog_item_products_catalog_item_id_idx on catalog_item_products (catalog_item_id);
--rollback drop table if exists catalog_item_products;
