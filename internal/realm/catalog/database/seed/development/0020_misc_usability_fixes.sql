--liquibase formatted sql

--changeset pixels:pixels-catalog-seed-development-0020-misc-usability-fixes context:development
update catalog_pages
set parent_id = 1,
    order_num = 26,
    visible = true,
    enabled = true,
    updated_at = now()
where id = 900;

update catalog_items
set name = case id
        when 200902 then 'classic_trophy_gold'
        when 200903 then 'classic_trophy_silver'
        when 200904 then 'classic_trophy_bronze'
    end,
    updated_at = now()
where id between 200902 and 200904;

update catalog_items
set extra_data = '1',
    updated_at = now()
where id between 200609 and 200620;

--rollback update catalog_pages set parent_id=5,order_num=4,updated_at=now() where id=900; update catalog_items set name=case id when 200902 then 'prize1' when 200903 then 'prize2' when 200904 then 'prize3' end,updated_at=now() where id between 200902 and 200904; update catalog_items set extra_data='0',updated_at=now() where id between 200609 and 200620;
