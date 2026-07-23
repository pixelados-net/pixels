--liquibase formatted sql

--changeset pixels:pixels-furniture-0015-add-room-final-interactions
alter table furniture_items
    add column rental_owner_player_id bigint null references players(id),
    add column rental_expires_at timestamptz null,
    add column rental_price_credits integer null,
    add column stack_height_override_cm integer null;

alter table furniture_items
    add constraint furniture_items_rental_pair_chk check ((rental_owner_player_id is null) = (rental_expires_at is null)),
    add constraint furniture_items_rental_price_chk check (rental_price_credits is null or rental_price_credits >= 0),
    add constraint furniture_items_stack_height_override_chk check (stack_height_override_cm is null or stack_height_override_cm between 0 and 4000);

create table furniture_lovelocks (
    item_id bigint primary key references furniture_items(id),
    player_one_id bigint not null references players(id),
    player_two_id bigint null references players(id),
    sealed_at timestamptz null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table player_mysterybox_keys (
    player_id bigint primary key references players(id),
    box_color text not null default 'default',
    key_color text not null default 'default',
    updated_at timestamptz not null default now()
);

create index furniture_items_active_rental_idx
    on furniture_items (rental_expires_at)
    where rental_owner_player_id is not null and deleted_at is null;

--rollback drop table if exists player_mysterybox_keys; drop table if exists furniture_lovelocks; alter table furniture_items drop column if exists stack_height_override_cm, drop column if exists rental_price_credits, drop column if exists rental_expires_at, drop column if exists rental_owner_player_id;
