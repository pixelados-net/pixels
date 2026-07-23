--liquibase formatted sql

--changeset pixels:pixels-furniture-0014-clothing-products
create table clothing_products (
    product_code text primary key,
    definition_id bigint not null unique references furniture_definitions(id),
    enabled boolean not null default true,
    constraint clothing_products_code_chk check (char_length(product_code) between 1 and 64)
);

create table clothing_product_sets (
    product_code text not null references clothing_products(product_code) on delete cascade,
    figure_set_id integer not null check (figure_set_id > 0),
    primary key (product_code, figure_set_id)
);

create table player_clothing_sets (
    player_id bigint not null references players(id) on delete cascade,
    figure_set_id integer not null check (figure_set_id > 0),
    product_code text not null references clothing_products(product_code),
    unlocked_at timestamptz not null default now(),
    primary key (player_id, figure_set_id)
);

create index player_clothing_sets_player_idx on player_clothing_sets(player_id, figure_set_id);

--rollback drop index if exists player_clothing_sets_player_idx; drop table if exists player_clothing_sets; drop table if exists clothing_product_sets; drop table if exists clothing_products;
