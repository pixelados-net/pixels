--liquibase formatted sql

--changeset pixels:pixels-catalog-seed-development-0026-group-page-link-compatibility context:development
update catalog_pages
set name = 'guild_custom_furni',
    updated_at = now()
where id = 116
  and name = 'group_furniture';
--rollback update catalog_pages set name='group_furniture',updated_at=now() where id=116 and name='guild_custom_furni';
