--liquibase formatted sql
--changeset pixels:pixels-pet-0002-create-pets
create table pets (
    id bigint generated always as identity primary key,
    owner_player_id bigint not null references players(id),
    name text not null check (char_length(name) between 2 and 16),
    type_id integer not null references pet_species(type_id),
    breed_id integer not null,
    palette_id integer not null,
    color text not null check (color ~ '^[0-9A-F]{6}$'),
    rarity integer not null default 0 check (rarity >= 0),
    level integer not null default 1 check (level between 1 and 20),
    experience integer not null default 0 check (experience >= 0),
    energy integer not null default 100 check (energy >= 0),
    happiness integer not null default 100 check (happiness between 0 and 100),
    respect integer not null default 0 check (respect >= 0),
    stats_at timestamptz not null default now(),
    room_id bigint null references rooms(id) on delete cascade,
    x integer null,
    y integer null,
    z double precision null,
    rotation smallint null,
    posture text not null default 'std',
    has_saddle boolean not null default false,
    can_breed boolean not null default true,
    public_ride boolean not null default false,
    public_breed boolean not null default false,
    grow_at timestamptz null,
    die_at timestamptz null,
    state text not null default 'inventory' check (state in ('inventory','room','breeding_reserved','harvested','composted')),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz null,
    version bigint not null default 1,
    foreign key (type_id,breed_id,palette_id) references pet_breeds(type_id,breed_id,palette_id),
    constraint pets_placement_chk check (
        (room_id is null and x is null and y is null and z is null and rotation is null and state <> 'room') or
        (room_id is not null and x is not null and y is not null and z is not null and rotation between 0 and 7 and state = 'room')
    )
);
create index pets_owner_inventory_idx on pets(owner_player_id,id) where deleted_at is null and room_id is null;
create index pets_room_idx on pets(room_id,id) where deleted_at is null and room_id is not null;
create index pets_lifecycle_idx on pets(least(grow_at,die_at)) where deleted_at is null and type_id=16;

create table pet_appearance_parts (
    pet_id bigint not null references pets(id) on delete cascade,
    ordinal integer not null check (ordinal between 0 and 31),
    layer_id integer not null,
    part_id integer not null,
    palette_id integer not null,
    primary key (pet_id,ordinal)
);

create table pet_respects (
    pet_id bigint not null references pets(id) on delete cascade,
    actor_player_id bigint not null references players(id) on delete cascade,
    respected_on date not null,
    primary key (pet_id,actor_player_id,respected_on)
);
--rollback drop table if exists pet_respects; drop table if exists pet_appearance_parts; drop table if exists pets;
