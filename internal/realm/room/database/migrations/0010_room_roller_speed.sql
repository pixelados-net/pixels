--liquibase formatted sql

--changeset pixels:pixels-room-0010-room-roller-speed
alter table rooms add column roller_speed integer not null default 4;
alter table rooms add constraint rooms_roller_speed_chk check (roller_speed >= -1 and roller_speed <= 20);
--rollback alter table rooms drop constraint if exists rooms_roller_speed_chk;
--rollback alter table rooms drop column if exists roller_speed;
