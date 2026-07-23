--liquibase formatted sql

--changeset pixels:pixels-catalog-0002-create-catalog-items
create table catalog_items (
    id bigint generated always as identity primary key,
    page_id bigint not null references catalog_pages(id),
    definition_id bigint not null references furniture_definitions(id),
    name text not null,
    cost_credits bigint not null default 0,
    cost_points bigint not null default 0,
    points_type integer not null default -1,
    amount integer not null default 1,
    limited_stack integer not null default 0,
    limited_sells integer not null default 0,
    offer_id bigint null,
    club_only boolean not null default false,
    order_num integer not null default 0,
    enabled boolean not null default true,
    extra_data text not null default '',
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz null,
    version bigint not null default 1,
    constraint catalog_items_name_length_chk check (char_length(name) between 1 and 64),
    constraint catalog_items_pricing_non_negative_chk check (cost_credits >= 0 and cost_points >= 0),
    constraint catalog_items_amount_positive_chk check (amount > 0),
    constraint catalog_items_limited_non_negative_chk check (limited_stack >= 0 and limited_sells >= 0),
    constraint catalog_items_limited_sells_bound_chk check (limited_stack = 0 or limited_sells <= limited_stack),
    constraint catalog_items_version_positive_chk check (version > 0)
);

create index catalog_items_page_id_idx on catalog_items (page_id) where deleted_at is null;
create index catalog_items_definition_id_idx on catalog_items (definition_id) where deleted_at is null;
--rollback drop table if exists catalog_items;
