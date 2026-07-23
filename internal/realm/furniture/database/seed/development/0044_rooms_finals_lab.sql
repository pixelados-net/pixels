--liquibase formatted sql

--changeset pixels:furniture-seed-rooms-finals-0044 context:development
insert into furniture_definitions(
 id,sprite_id,name,public_name,kind,width,length,stack_height,
 allow_stack,allow_walk,allow_sit,allow_lay,allow_inventory_stack,
 interaction_type,interaction_modes_count,multiheight,custom_params,metadata
)
overriding system value values
 (970001,244,'qa_lovelock','QA Lovelock','floor',1,1,0,false,false,false,false,true,'lovelock',2,'','','{"qa":"rooms-finals"}'),
 (970002,239,'qa_mystery_box','QA Mystery Box','floor',1,1,1,true,false,false,false,true,'mystery_box',2,'','','{"qa":"rooms-finals"}'),
 (970003,5315,'qa_mystery_trophy','QA Mystery Trophy','floor',1,1,1,true,false,false,false,true,'mystery_trophy',1,'','','{"qa":"rooms-finals"}'),
 (970004,4744,'qa_firework','QA Firework','floor',1,1,1,true,false,false,false,true,'firework',3,'','5','{"qa":"rooms-finals"}'),
 (970005,3632,'qa_stack_helper','QA Stack Helper','floor',1,1,0,true,true,false,false,true,'custom_stack_height',1,'','','{"qa":"rooms-finals"}'),
 (970006,1649,'qa_rentable','QA Rentable Space','floor',1,1,0.5,true,true,false,false,true,'rentable_space',1,'','','{"qa":"rooms-finals"}')
on conflict(id) do update set sprite_id=excluded.sprite_id,name=excluded.name,public_name=excluded.public_name,
 interaction_type=excluded.interaction_type,interaction_modes_count=excluded.interaction_modes_count,
 custom_params=excluded.custom_params,deleted_at=null,updated_at=now();

insert into furniture_items(id,definition_id,owner_player_id,room_id,x,y,z,rotation,wall_position,extra_data,rental_price_credits)
overriding system value values
 (970101,13,1,160,2,2,0,0,null,'0',null),
 (970102,31,1,160,4,2,0,0,null,'0',null),
 (970103,20,1,160,null,null,null,0,':w=4,2 l=1,1 r','0',null),
 (970104,970001,1,160,6,2,0,0,null,'0',null),
 (970105,970002,1,160,8,2,0,0,null,'0',null),
 (970106,970003,1,160,10,2,0,0,null,'0',null),
 (970107,970004,1,160,4,5,0,0,null,'1',null),
 (970108,970005,1,160,6,5,0,0,null,'0',null),
 (970109,970006,1,160,8,5,0,0,null,'0',10)
on conflict(id) do update set definition_id=excluded.definition_id,owner_player_id=excluded.owner_player_id,
 room_id=excluded.room_id,x=excluded.x,y=excluded.y,z=excluded.z,rotation=excluded.rotation,
 wall_position=excluded.wall_position,extra_data=excluded.extra_data,rental_price_credits=excluded.rental_price_credits,
 deleted_at=null,updated_at=now();

insert into player_mysterybox_keys(player_id,box_color,key_color)
select id,'blue','gold' from players where lower(username)='demo'
on conflict(player_id) do update set box_color=excluded.box_color,key_color=excluded.key_color,updated_at=now();

select setval(pg_get_serial_sequence('furniture_definitions','id'),greatest((select max(id) from furniture_definitions),1));
select setval(pg_get_serial_sequence('furniture_items','id'),greatest((select max(id) from furniture_items),1));
--rollback delete from player_mysterybox_keys where player_id in (select id from players where lower(username)='demo'); delete from furniture_items where id between 970101 and 970109; delete from furniture_definitions where id between 970001 and 970006;
