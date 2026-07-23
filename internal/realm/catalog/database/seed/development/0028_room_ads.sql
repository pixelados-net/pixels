--liquibase formatted sql

--changeset pixels:catalog-seed-room-ads-0028 context:development
insert into catalog_pages(id,parent_id,name,layout,icon_color,icon_image,min_rank,order_num,visible,enabled,club_only)
overriding system value values
 (990,1,'room_ads','room_ads',1,1,1,99,true,true,false)
on conflict(id) do update set parent_id=excluded.parent_id,name=excluded.name,layout=excluded.layout,
 visible=true,enabled=true,deleted_at=null,updated_at=now();

insert into catalog_items(id,page_id,definition_id,reward_kind,name,cost_credits,cost_points,points_type,amount,limited_stack,limited_sells,club_only,order_num,enabled,extra_data)
overriding system value values
 (990001,990,null,'service','room_ad_two_hours',10,0,-1,1,0,0,false,1,true,'')
on conflict(id) do update set page_id=excluded.page_id,reward_kind=excluded.reward_kind,name=excluded.name,
 cost_credits=excluded.cost_credits,enabled=true,deleted_at=null,updated_at=now();

select setval(pg_get_serial_sequence('catalog_pages','id'),greatest((select max(id) from catalog_pages),1));
select setval(pg_get_serial_sequence('catalog_items','id'),greatest((select max(id) from catalog_items),1));
--rollback delete from catalog_items where id=990001; delete from catalog_pages where id=990;
