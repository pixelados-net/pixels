--liquibase formatted sql

--changeset pixels:pixels-room-0014-add-hide-wired
alter table rooms
    add column hide_wired boolean not null default false;

--rollback alter table rooms drop column hide_wired;
