--liquibase formatted sql

--changeset pixels:pixels-catalog-0014-add-effect-grants
alter table catalog_items add column grants_effect_id integer null;
alter table catalog_items add column grants_effect_duration_seconds integer not null default 0;
alter table catalog_items drop constraint catalog_items_product_kind_chk;
alter table catalog_items add constraint catalog_items_effect_positive_chk check (grants_effect_id is null or grants_effect_id > 0);
alter table catalog_items add constraint catalog_items_effect_duration_chk check (grants_effect_duration_seconds >= 0);
alter table catalog_items add constraint catalog_items_product_kind_chk check (
    (room_bundle_template_room_id is null and (definition_id is not null or grants_effect_id is not null))
    or
    (room_bundle_template_room_id is not null and definition_id is null and grants_effect_id is null and amount = 0 and limited_stack = 0 and limited_sells = 0 and not giftable and not bundle_discount_enabled)
);

--rollback alter table catalog_items drop constraint if exists catalog_items_product_kind_chk;
--rollback alter table catalog_items drop constraint if exists catalog_items_effect_duration_chk;
--rollback alter table catalog_items drop constraint if exists catalog_items_effect_positive_chk;
--rollback alter table catalog_items drop column if exists grants_effect_duration_seconds;
--rollback alter table catalog_items drop column if exists grants_effect_id;
--rollback alter table catalog_items add constraint catalog_items_product_kind_chk check ((room_bundle_template_room_id is null and definition_id is not null) or (room_bundle_template_room_id is not null and definition_id is null and amount = 0 and limited_stack = 0 and limited_sells = 0 and not giftable and not bundle_discount_enabled));
