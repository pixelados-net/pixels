--liquibase formatted sql

--changeset pixels:pixels-catalog-0010-add-freshness
alter table catalog_pages add column expires_at timestamptz null;
alter table catalog_pages add column excluded_from_kickback boolean not null default false;
alter table catalog_items add column scheduled_at timestamptz null;
create table catalog_purchase_log (
    id bigint generated always as identity primary key,
    player_id bigint not null references players(id),
    catalog_item_id bigint not null references catalog_items(id),
    quantity integer not null,
    cost_credits bigint not null default 0,
    cost_points bigint not null default 0,
    points_type integer not null default -1,
    purchased_at timestamptz not null default now()
);
create index catalog_purchase_log_player_time_idx on catalog_purchase_log (player_id, purchased_at);
--rollback drop table if exists catalog_purchase_log; alter table catalog_items drop column scheduled_at; alter table catalog_pages drop column excluded_from_kickback; alter table catalog_pages drop column expires_at;
