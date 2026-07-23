--liquibase formatted sql

--changeset pixels:pixels-furniture-0003-allow-wall-definitions
alter table furniture_definitions
drop constraint furniture_definitions_kind_chk;

alter table furniture_definitions
add constraint furniture_definitions_kind_chk
check (kind in ('floor', 'wall'));
--rollback alter table furniture_definitions drop constraint furniture_definitions_kind_chk;
--rollback alter table furniture_definitions add constraint furniture_definitions_kind_chk check (kind = 'floor');
