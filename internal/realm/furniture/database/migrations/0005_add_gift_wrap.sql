--liquibase formatted sql

--changeset pixels:pixels-furniture-0005-add-gift-wrap
alter table furniture_items
    add column gift_wrapped boolean not null default false,
    add column gift_wrap_box_id integer null,
    add column gift_wrap_ribbon_id integer null,
    add column gift_sender_player_id bigint null references players(id),
    add column gift_message text null,
    add constraint furniture_items_gift_message_length_chk check (gift_message is null or char_length(gift_message) <= 255);
--rollback alter table furniture_items drop constraint if exists furniture_items_gift_message_length_chk; alter table furniture_items drop column if exists gift_message, drop column if exists gift_sender_player_id, drop column if exists gift_wrap_ribbon_id, drop column if exists gift_wrap_box_id, drop column if exists gift_wrapped;
