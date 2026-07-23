--liquibase formatted sql

--changeset pixels:pixels-furniture-0008-add-effect-fields
alter table furniture_definitions add column effect_pool integer[] null;
alter table furniture_definitions add column effect_male integer null;
alter table furniture_definitions add column effect_female integer null;
alter table furniture_definitions add constraint furniture_effect_pool_positive_chk check (effect_pool is null or 0 < all(effect_pool));
alter table furniture_definitions add constraint furniture_effect_gender_positive_chk check ((effect_male is null or effect_male > 0) and (effect_female is null or effect_female > 0));

--rollback alter table furniture_definitions drop constraint if exists furniture_effect_gender_positive_chk;
--rollback alter table furniture_definitions drop constraint if exists furniture_effect_pool_positive_chk;
--rollback alter table furniture_definitions drop column if exists effect_female;
--rollback alter table furniture_definitions drop column if exists effect_male;
--rollback alter table furniture_definitions drop column if exists effect_pool;
