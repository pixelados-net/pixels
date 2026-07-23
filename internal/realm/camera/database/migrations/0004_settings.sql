--liquibase formatted sql

--changeset pixels:pixels-camera-0004-settings
create table camera_settings (
    id smallint primary key default 1 check (id = 1),
    enabled boolean not null default true,
    credits_price bigint not null default 2 check (credits_price >= 0),
    points_price bigint not null default 0 check (points_price >= 0),
    points_type integer not null default 5,
    publish_points_price bigint not null default 10 check (publish_points_price >= 0),
    publish_points_type integer not null default 5,
    publish_cooldown_seconds integer not null default 180 check (publish_cooldown_seconds >= 0),
    updated_at timestamptz not null default now(),
    version bigint not null default 1 check (version > 0)
);
insert into camera_settings (id) values (1) on conflict (id) do nothing;
--rollback drop table if exists camera_settings;
