--liquibase formatted sql

--changeset pixels:pixels-catalog-0005-bundle-support
alter table catalog_items drop column offer_id;
alter table catalog_items add column bundle_discount_enabled boolean not null default false;
alter table catalog_items drop constraint catalog_items_amount_positive_chk;
alter table catalog_items add constraint catalog_items_amount_non_negative_chk check (amount >= 0);
--rollback alter table catalog_items drop constraint catalog_items_amount_non_negative_chk; alter table catalog_items add constraint catalog_items_amount_positive_chk check (amount > 0); alter table catalog_items drop column bundle_discount_enabled; alter table catalog_items add column offer_id bigint null;
