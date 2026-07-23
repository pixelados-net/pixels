--liquibase formatted sql

--changeset pixels:pixels-catalog-0011-link-purchase-items
create table catalog_purchase_items (
    purchase_id bigint not null references catalog_purchase_log(id) on delete cascade,
    furniture_item_id bigint not null references furniture_items(id),
    primary key (purchase_id, furniture_item_id),
    unique (furniture_item_id)
);
--rollback drop table if exists catalog_purchase_items;
