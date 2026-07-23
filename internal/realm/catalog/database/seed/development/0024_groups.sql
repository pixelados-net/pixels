--liquibase formatted sql

--changeset pixels:pixels-catalog-seed-development-0024-groups context:development
insert into catalog_pages(id,parent_id,name,layout,icon_color,icon_image,order_num,visible,enabled,club_only,new_additions)
overriding system value values
    (115,1,'groups','guilds',1,203,10,true,true,false,true),
    (116,115,'group_furniture','guild_furni',1,28,1,true,true,false,false),
    (117,115,'group_forums','guild_forum',1,207,2,true,true,false,false)
on conflict(id) do update set parent_id=excluded.parent_id,name=excluded.name,layout=excluded.layout,
    icon_color=excluded.icon_color,icon_image=excluded.icon_image,order_num=excluded.order_num,
    visible=true,enabled=true,club_only=false,new_additions=excluded.new_additions,deleted_at=null,updated_at=now();

insert into catalog_items(id,page_id,definition_id,name,cost_credits,cost_points,points_type,amount,limited_stack,limited_sells,bundle_discount_enabled,giftable,club_only,order_num,enabled,extra_data,reward_kind)
overriding system value values
    (1401,116,5374,'group_custom_furniture',4,0,-1,1,0,0,false,false,false,1,true,'0','furniture'),
    (1402,117,5863,'group_forum_terminal',15,0,-1,1,0,0,false,false,false,1,true,'0','furniture')
on conflict(id) do update set page_id=excluded.page_id,definition_id=excluded.definition_id,name=excluded.name,
    cost_credits=excluded.cost_credits,cost_points=excluded.cost_points,points_type=excluded.points_type,
    amount=1,limited_stack=0,limited_sells=0,bundle_discount_enabled=false,giftable=false,
    club_only=false,order_num=excluded.order_num,enabled=true,extra_data=excluded.extra_data,
    reward_kind=excluded.reward_kind,deleted_at=null,updated_at=now();

select setval(pg_get_serial_sequence('catalog_pages','id'),greatest((select max(id) from catalog_pages),1));
select setval(pg_get_serial_sequence('catalog_items','id'),greatest((select max(id) from catalog_items),1));
--rollback delete from catalog_items where id in (1401,1402); delete from catalog_pages where id between 115 and 117;
