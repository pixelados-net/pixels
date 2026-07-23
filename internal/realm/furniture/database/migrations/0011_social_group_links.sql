--liquibase formatted sql

--changeset pixels:pixels-furniture-0011-social-group-links
create table furniture_social_group_links (
    item_id bigint primary key references furniture_items(id) on delete cascade,
    group_id bigint null references social_groups(id) on delete set null,
    linked_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    version bigint not null default 1,
    constraint furniture_social_group_links_version_chk check(version > 0)
);

create index furniture_social_group_links_group_idx on furniture_social_group_links(group_id) where group_id is not null;

--rollback drop table if exists furniture_social_group_links;
