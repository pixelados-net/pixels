--liquibase formatted sql

--changeset pixels:pixels-furniture-0013-add-description
alter table furniture_definitions
    add column description text not null default '';
--rollback alter table furniture_definitions drop column if exists description;
