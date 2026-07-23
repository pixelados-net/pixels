--liquibase formatted sql

--changeset pixels:pixels-catalog-0001-create-catalog-pages
create table catalog_pages (
    id bigint generated always as identity primary key,
    parent_id bigint null references catalog_pages(id),
    name text not null,
    layout text not null default 'default_3x3',
    icon_color integer not null default 0,
    icon_image integer not null default 0,
    min_rank integer not null default 1,
    order_num integer not null default 0,
    visible boolean not null default true,
    enabled boolean not null default true,
    club_only boolean not null default false,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz null,
    version bigint not null default 1,
    constraint catalog_pages_name_length_chk check (char_length(name) between 1 and 64),
    constraint catalog_pages_version_positive_chk check (version > 0)
);

create unique index catalog_pages_name_active_uidx on catalog_pages (name) where deleted_at is null;
create index catalog_pages_parent_id_idx on catalog_pages (parent_id) where deleted_at is null;
--rollback drop table if exists catalog_pages;
