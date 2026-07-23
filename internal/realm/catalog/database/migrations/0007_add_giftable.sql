--liquibase formatted sql

--changeset pixels:pixels-catalog-0007-add-giftable
alter table catalog_items add column giftable boolean not null default false;
--rollback alter table catalog_items drop column giftable;
