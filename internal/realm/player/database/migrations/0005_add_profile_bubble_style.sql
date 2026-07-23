--liquibase formatted sql

--changeset pixels:player-0005
alter table player_profiles
    add column bubble_style integer not null default 0;

--rollback alter table player_profiles drop column bubble_style;
