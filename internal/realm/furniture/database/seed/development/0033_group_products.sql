--liquibase formatted sql

--changeset pixels:pixels-furniture-seed-development-0033-group-products context:development
-- Legacy Arcturus identifiers are retained because Nitro's production assets expose both sprites.
insert into furniture_definitions(id,sprite_id,name,public_name,kind,width,length,stack_height,allow_stack,allow_walk,allow_sit,allow_lay,allow_inventory_stack,allow_trade,allow_marketplace_sale,interaction_type,interaction_modes_count,multiheight,custom_params,metadata)
overriding system value values
    (5374,5374,'guild_customized','Custom Group Furniture','floor',1,1,1.00,true,false,false,false,true,true,true,'group_furniture',2,'','','{"socialGroup":true}'),
    (5863,5863,'guild_forum','Group Forum Terminal','floor',2,1,0.00,true,false,false,false,true,true,false,'group_forum',1,'','','{"socialGroup":true,"forumEntitlement":true}')
on conflict(id) do update set sprite_id=excluded.sprite_id,name=excluded.name,public_name=excluded.public_name,
    kind=excluded.kind,width=excluded.width,length=excluded.length,stack_height=excluded.stack_height,
    allow_stack=excluded.allow_stack,allow_walk=excluded.allow_walk,allow_inventory_stack=excluded.allow_inventory_stack,
    allow_trade=excluded.allow_trade,allow_marketplace_sale=excluded.allow_marketplace_sale,
    interaction_type=excluded.interaction_type,interaction_modes_count=excluded.interaction_modes_count,
    metadata=excluded.metadata,deleted_at=null,updated_at=now();

insert into furniture_items(id,definition_id,owner_player_id,room_id,x,y,z,rotation,extra_data)
overriding system value values
    (430001,5374,1,131,4,2,0,0,'0'),
    (430002,5374,3,135,4,2,0,0,'0'),
    (430003,5374,2,135,6,2,0,0,'1'),
    (430004,5863,1,134,4,2,0,0,'0')
on conflict(id) do update set definition_id=excluded.definition_id,owner_player_id=excluded.owner_player_id,
    room_id=excluded.room_id,x=excluded.x,y=excluded.y,z=excluded.z,rotation=excluded.rotation,
    extra_data=excluded.extra_data,deleted_at=null,updated_at=now();

insert into furniture_social_group_links(item_id,group_id) values
    (430001,2),(430002,2),(430003,2),(430004,5)
on conflict(item_id) do update set group_id=excluded.group_id,updated_at=now();

select setval(pg_get_serial_sequence('furniture_definitions','id'),greatest((select max(id) from furniture_definitions),1));
select setval(pg_get_serial_sequence('furniture_items','id'),greatest((select max(id) from furniture_items),1));
--rollback delete from furniture_items where id between 430001 and 430004; delete from furniture_definitions where id in (5374,5863);
