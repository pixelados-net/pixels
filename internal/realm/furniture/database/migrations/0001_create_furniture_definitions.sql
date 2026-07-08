--liquibase formatted sql

--changeset pixels:pixels-furniture-0001-create-furniture-definitions
create table furniture_definitions (
    id bigint generated always as identity primary key,
    sprite_id integer not null,
    name text not null,
    public_name text not null,
    kind text not null default 'floor',
    width smallint not null default 1,
    length smallint not null default 1,
    stack_height numeric(4,2) not null default 0,
    allow_stack boolean not null default true,
    allow_walk boolean not null default false,
    allow_sit boolean not null default false,
    allow_lay boolean not null default false,
    allow_inventory_stack boolean not null default true,
    interaction_type text not null default 'default',
    interaction_modes_count integer not null default 1,
    multiheight text not null default '',
    custom_params text not null default '',
    metadata jsonb not null default '{}'::jsonb,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz null,
    version bigint not null default 1,
    constraint furniture_definitions_name_length_chk check (char_length(name) between 1 and 70),
    constraint furniture_definitions_public_name_length_chk check (char_length(public_name) between 1 and 56),
    constraint furniture_definitions_kind_chk check (kind = 'floor'),
    constraint furniture_definitions_width_positive_chk check (width > 0),
    constraint furniture_definitions_length_positive_chk check (length > 0),
    constraint furniture_definitions_stack_height_positive_chk check (stack_height >= 0),
    constraint furniture_definitions_interaction_modes_count_positive_chk check (interaction_modes_count > 0),
    constraint furniture_definitions_version_positive_chk check (version > 0)
);

create unique index furniture_definitions_name_active_uidx
on furniture_definitions (name)
where deleted_at is null;
--rollback drop table if exists furniture_definitions;
