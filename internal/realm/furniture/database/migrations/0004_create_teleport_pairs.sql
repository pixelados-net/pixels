--liquibase formatted sql

--changeset pixels:pixels-furniture-0004-create-teleport-pairs
create table furniture_item_teleport_pairs (
    item_one_id bigint primary key references furniture_items(id) on delete cascade,
    item_two_id bigint not null unique references furniture_items(id) on delete cascade,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    constraint furniture_item_teleport_pairs_order_chk check (item_one_id < item_two_id)
);

create index furniture_item_teleport_pairs_item_two_idx on furniture_item_teleport_pairs (item_two_id);
--rollback drop table if exists furniture_item_teleport_pairs;
