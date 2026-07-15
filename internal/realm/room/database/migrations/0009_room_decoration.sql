--liquibase formatted sql

--changeset pixels:pixels-room-0009-room-decoration
alter table rooms add column floor_paint varchar(32) not null default '0.0';
alter table rooms add column wallpaper varchar(32) not null default '0.0';
alter table rooms add column landscape varchar(32) not null default '0.0';
create table room_dimmer_presets (
    room_id bigint not null references rooms(id) on delete cascade,
    preset_id smallint not null check (preset_id between 1 and 3),
    background_only boolean not null default false,
    color char(7) not null default '#000000',
    brightness smallint not null default 255 check (brightness between 76 and 255),
    selected boolean not null default false,
    enabled boolean not null default false,
    updated_at timestamptz not null default now(),
    primary key (room_id, preset_id)
);
create unique index room_dimmer_selected_idx on room_dimmer_presets(room_id) where selected;
--rollback drop table if exists room_dimmer_presets; alter table rooms drop column if exists landscape; alter table rooms drop column if exists wallpaper; alter table rooms drop column if exists floor_paint;
