--liquibase formatted sql

--changeset pixels:pixels-furniture-0006-add-gift-wrap-sprite
alter table furniture_items
    add column gift_wrap_sprite_id integer null,
    add constraint furniture_items_gift_wrap_sprite_positive_chk
        check (gift_wrap_sprite_id is null or gift_wrap_sprite_id > 0);
update furniture_items set gift_wrap_sprite_id = 187 where gift_wrapped = true and gift_wrap_sprite_id is null;
--rollback alter table furniture_items drop constraint if exists furniture_items_gift_wrap_sprite_positive_chk; alter table furniture_items drop column if exists gift_wrap_sprite_id;
