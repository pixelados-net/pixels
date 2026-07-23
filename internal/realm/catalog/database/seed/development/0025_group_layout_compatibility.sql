--liquibase formatted sql

--changeset pixels:pixels-catalog-seed-development-0025-group-layout-compatibility context:development
update catalog_pages
set layout = case id
        when 115 then 'guild_frontpage'
        when 116 then 'guild_custom_furni'
        else layout
    end,
    updated_at = now()
where (id = 115 and name = 'groups')
   or (id = 116 and name = 'group_furniture');
--rollback update catalog_pages set layout=case id when 115 then 'guilds' when 116 then 'guild_furni' else layout end,updated_at=now() where id in (115,116);
