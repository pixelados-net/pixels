--liquibase formatted sql
--changeset pixels:pixels-furniture-seed-development-0034-crafting-lab context:development
insert into furniture_definitions(id,sprite_id,name,public_name,kind,width,length,stack_height,allow_stack,allow_walk,allow_sit,allow_lay,allow_inventory_stack,allow_trade,allow_marketplace_sale,allow_recycle,interaction_type,interaction_modes_count,multiheight,custom_params,metadata)
overriding system value values
 (3095,3095,'ecotron_box','Ecotron','floor',1,1,1,false,false,false,false,true,true,false,false,'default',1,'','','{}'),
 (8388,8388,'hween_c15_altar','Crafting Altar','floor',2,1,1,false,false,false,false,true,true,false,false,'default',1,'','','{}')
on conflict(id) do update set sprite_id=excluded.sprite_id,name=excluded.name,public_name=excluded.public_name,allow_recycle=excluded.allow_recycle,deleted_at=null,updated_at=now();

update furniture_definitions set allow_recycle=true,updated_at=now() where id in(1,2,3);
insert into furniture_items(id,definition_id,owner_player_id,room_id,x,y,z,rotation) overriding system value values
 (910001,8388,1,1,5,12,0,0),(910002,3095,1,1,8,12,0,0)
on conflict(id) do update set deleted_at=null,room_id=excluded.room_id,x=excluded.x,y=excluded.y,z=excluded.z,owner_player_id=excluded.owner_player_id;
insert into furniture_items(id,definition_id,owner_player_id,extra_data) overriding system value
select 910100+n,case when n%3=0 then 1 when n%3=1 then 2 else 3 end,1,'0' from generate_series(1,24)n
on conflict(id) do update set deleted_at=null,room_id=null,x=null,y=null,z=null,owner_player_id=1,marketplace_reserved=false;
select setval(pg_get_serial_sequence('furniture_definitions','id'),greatest((select max(id) from furniture_definitions),1));
select setval(pg_get_serial_sequence('furniture_items','id'),greatest((select max(id) from furniture_items),1));
--rollback delete from furniture_items where id between 910001 and 910124; update furniture_definitions set allow_recycle=false where id in(1,2,3); delete from furniture_definitions where id in(3095,8388);
