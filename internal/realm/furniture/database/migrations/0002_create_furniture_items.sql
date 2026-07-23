--liquibase formatted sql

--changeset pixels:pixels-furniture-0002-create-furniture-items
create table furniture_items (
    id bigint generated always as identity primary key,
    definition_id bigint not null references furniture_definitions(id),
    owner_player_id bigint not null references players(id),
    room_id bigint null references rooms(id),
    x smallint null,
    y smallint null,
    z numeric(6,2) null,
    rotation smallint not null default 0,
    wall_position text null,
    extra_data text not null default '0',
    metadata jsonb not null default '{}'::jsonb,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz null,
    version bigint not null default 1,
    constraint furniture_items_rotation_chk check (rotation in (0, 2, 4, 6)),
    constraint furniture_items_room_placement_chk check (
        (room_id is null and x is null and y is null and z is null)
        or
        (room_id is not null and x is not null and y is not null and z is not null)
    ),
    constraint furniture_items_version_positive_chk check (version > 0)
);

create index furniture_items_room_id_idx on furniture_items (room_id) where deleted_at is null;
create index furniture_items_owner_player_id_idx on furniture_items (owner_player_id) where deleted_at is null;
create index furniture_items_definition_id_idx on furniture_items (definition_id) where deleted_at is null;
--rollback drop table if exists furniture_items;
