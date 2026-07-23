--liquibase formatted sql

--changeset pixels:pixels-catalog-0015-room-bundle-bot-audit
alter table room_bundle_purchases add column bot_count integer not null default 0;
alter table room_bundle_purchases add constraint room_bundle_purchases_bot_count_chk check (bot_count >= 0);
--rollback alter table room_bundle_purchases drop constraint if exists room_bundle_purchases_bot_count_chk; alter table room_bundle_purchases drop column if exists bot_count;
