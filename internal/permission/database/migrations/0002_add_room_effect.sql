--liquibase formatted sql

--changeset pixels:pixels-permission-0002-add-room-effect
alter table permission_groups add column room_effect_id integer null;
alter table permission_groups add constraint permission_groups_room_effect_positive_chk check (room_effect_id is null or room_effect_id > 0);

--rollback alter table permission_groups drop constraint if exists permission_groups_room_effect_positive_chk;
--rollback alter table permission_groups drop column if exists room_effect_id;
