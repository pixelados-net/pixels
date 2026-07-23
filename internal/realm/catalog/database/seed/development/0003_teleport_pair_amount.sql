--liquibase formatted sql

--changeset pixels:pixels-catalog-seed-development-0003-teleport-pair-amount context:development
update catalog_items
set amount = 2, updated_at = now()
where id = 14 and definition_id = 8;
--rollback update catalog_items set amount = 1, updated_at = now() where id = 14 and definition_id = 8;
