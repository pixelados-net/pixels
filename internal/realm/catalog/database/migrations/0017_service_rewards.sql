--liquibase formatted sql

--changeset pixels:pixels-catalog-0017-service-rewards
alter table catalog_items drop constraint catalog_items_reward_kind_chk;
alter table catalog_items drop constraint catalog_items_product_kind_chk;
alter table catalog_items add constraint catalog_items_reward_kind_chk check (reward_kind in ('furniture','pet','service'));
alter table catalog_items add constraint catalog_items_product_kind_chk check (
    (room_bundle_template_room_id is not null and reward_kind='furniture' and definition_id is null and grants_effect_id is null and pet_type_id is null and pet_product_code='' and amount=0 and limited_stack=0 and limited_sells=0 and not giftable and not bundle_discount_enabled)
    or
    (room_bundle_template_room_id is null and reward_kind='furniture' and definition_id is not null and grants_effect_id is null and pet_type_id is null and pet_product_code='')
    or
    (room_bundle_template_room_id is null and reward_kind='furniture' and definition_id is null and grants_effect_id is not null and pet_type_id is null and pet_product_code='')
    or
    (room_bundle_template_room_id is null and reward_kind='pet' and definition_id is null and grants_effect_id is null and pet_type_id between 0 and 35 and char_length(pet_product_code) between 1 and 64 and amount=1 and limited_stack=0 and limited_sells=0 and not giftable and not bundle_discount_enabled)
    or
    (room_bundle_template_room_id is null and reward_kind='service' and definition_id is null and grants_effect_id is null and pet_type_id is null and pet_product_code='' and amount=1 and limited_stack=0 and limited_sells=0 and not giftable and not bundle_discount_enabled)
);
--rollback alter table catalog_items drop constraint catalog_items_product_kind_chk; alter table catalog_items drop constraint catalog_items_reward_kind_chk; alter table catalog_items add constraint catalog_items_reward_kind_chk check (reward_kind in ('furniture','pet')); alter table catalog_items add constraint catalog_items_product_kind_chk check ((room_bundle_template_room_id is not null and reward_kind='furniture' and definition_id is null and grants_effect_id is null and pet_type_id is null and pet_product_code='' and amount=0 and limited_stack=0 and limited_sells=0 and not giftable and not bundle_discount_enabled) or (room_bundle_template_room_id is null and reward_kind='furniture' and definition_id is not null and grants_effect_id is null and pet_type_id is null and pet_product_code='') or (room_bundle_template_room_id is null and reward_kind='furniture' and definition_id is null and grants_effect_id is not null and pet_type_id is null and pet_product_code='') or (room_bundle_template_room_id is null and reward_kind='pet' and definition_id is null and grants_effect_id is null and pet_type_id between 0 and 35 and char_length(pet_product_code) between 1 and 64 and amount=1 and limited_stack=0 and limited_sells=0 and not giftable and not bundle_discount_enabled));
