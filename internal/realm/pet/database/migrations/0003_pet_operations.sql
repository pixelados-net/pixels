--liquibase formatted sql
--changeset pixels:pixels-pet-0003-operations
create table pet_operations (
    id bigint generated always as identity primary key,
    idempotency_key text not null unique,
    pet_id bigint not null references pets(id) on delete cascade,
    kind text not null check (kind in ('grant','package','breeding','harvest','compost')),
    state text not null check (state in ('pending','completed','failed')),
    result jsonb not null default '{}'::jsonb,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table pet_breeding_sessions (
    nest_item_id bigint primary key references furniture_items(id) on delete cascade,
    room_id bigint not null references rooms(id) on delete cascade,
    generation_token text not null,
    parent_one_id bigint not null references pets(id),
    parent_two_id bigint not null references pets(id),
    owner_one_confirmed boolean not null default false,
    owner_two_confirmed boolean not null default false,
    state text not null check (state in ('requested','confirmed','completed','cancelled')),
    expires_at timestamptz not null,
    version bigint not null default 1,
    constraint pet_breeding_distinct_parents_chk check (parent_one_id <> parent_two_id)
);
create index pet_breeding_expiry_idx on pet_breeding_sessions(expires_at) where state in ('requested','confirmed');

create table pet_audit_entries (
    id bigint generated always as identity primary key,
    pet_id bigint null references pets(id) on delete set null,
    actor_player_id bigint null references players(id) on delete set null,
    action text not null,
    detail jsonb not null default '{}'::jsonb,
    created_at timestamptz not null default now()
);
create index pet_audit_pet_idx on pet_audit_entries(pet_id,created_at desc);
--rollback drop table if exists pet_audit_entries; drop table if exists pet_breeding_sessions; drop table if exists pet_operations;
