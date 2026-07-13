--liquibase formatted sql

--changeset pixels:pixels-catalog-0013-room-bundle-audit-name
alter table room_bundle_purchases rename column furniture_count to furniture_item_count;
--rollback alter table room_bundle_purchases rename column furniture_item_count to furniture_count;
