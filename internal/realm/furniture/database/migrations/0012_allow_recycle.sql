--liquibase formatted sql

--changeset pixels:pixels-furniture-0012-allow-recycle
alter table furniture_definitions
    add column allow_recycle boolean not null default false;
--rollback alter table furniture_definitions drop column if exists allow_recycle;
