--liquibase formatted sql

--changeset pixels:pixels-catalog-0009-add-new-additions
alter table catalog_pages add column new_additions boolean not null default false;
create table catalog_new_additions_seen (
    player_id bigint primary key references players(id),
    last_seen_at timestamptz not null default now()
);
--rollback drop table if exists catalog_new_additions_seen; alter table catalog_pages drop column new_additions;
