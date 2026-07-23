--liquibase formatted sql

--changeset pixels:furniture-seed-games-qa-maps-0039 context:development
--validCheckSum:ANY
insert into furniture_definitions(id,sprite_id,name,public_name,kind,width,length,stack_height,allow_stack,allow_walk,allow_sit,allow_lay,allow_inventory_stack,interaction_type,interaction_modes_count,multiheight,custom_params,metadata)
overriding system value values
 (950030,3741,'es_rollerskate_tagging','Rollerskate Pole','floor',1,1,3.00,true,false,false,false,true,'rollerskate_pole',2,'','','{"game":"tag","seed":"games-showcase"}')
on conflict(id) do update set sprite_id=excluded.sprite_id,name=excluded.name,public_name=excluded.public_name,
 interaction_type=excluded.interaction_type,interaction_modes_count=excluded.interaction_modes_count,
 metadata=excluded.metadata,deleted_at=null,updated_at=now();

delete from room_wired_settings where item_id between 960000 and 960499 or item_id between 961000 and 961599;
delete from furniture_items where id between 960000 and 960499 or id between 961000 and 961599;

insert into furniture_items(id,definition_id,owner_player_id,room_id,x,y,z,rotation,extra_data)
overriding system value
select 961000+(y-4)*8+(x-5),950001,1,150,x,y,0,0,'0'
from generate_series(5,12) x cross join generate_series(4,11) y;

insert into furniture_items(id,definition_id,owner_player_id,room_id,x,y,z,rotation,extra_data)
overriding system value values
 (961100,950002,1,150,4,4,0,0,'0'),(961101,950003,1,150,13,4,0,0,'0'),
 (961102,950004,1,150,4,11,0,0,'0'),(961103,950005,1,150,13,11,0,0,'0'),
 (961104,950006,1,150,2,5,0,0,'30'),(961105,950007,1,150,15,7,0,0,'0'),
 (961106,950008,1,150,8,7,0,0,'0'),(961107,950009,1,150,3,13,0,0,'0'),
 (961108,950009,1,150,14,13,0,0,'0'),(961109,900010,1,150,5,14,0,0,'0'),
 (961110,900011,1,150,6,14,0,0,'0'),(961111,900012,1,150,7,14,0,0,'0');

insert into room_wired_settings(item_id,int_params,string_param,selection_mode,delay_pulses) values
 (961109,'[]','BattleBallPlayer',0,0),(961110,'[]','900001',0,0),(961111,'[]','900001',0,0)
on conflict(item_id) do update set int_params=excluded.int_params,string_param=excluded.string_param,
 selection_mode=excluded.selection_mode,delay_pulses=excluded.delay_pulses,updated_at=now();

insert into furniture_items(id,definition_id,owner_player_id,room_id,x,y,z,rotation,extra_data)
overriding system value
select 961200+(y-4)*9+(x-5),950010,1,151,x,y,0,0,'0'
from generate_series(5,13) x cross join generate_series(4,10) y;

insert into furniture_items(id,definition_id,owner_player_id,room_id,x,y,z,rotation,extra_data)
overriding system value values
 (961300,950011,1,151,7,4,0.01,0,'0'),(961301,950011,1,151,11,4,0.01,0,'0'),
 (961302,950011,1,151,6,5,0.01,0,'0'),(961303,950011,1,151,9,5,0.01,0,'0'),
 (961304,950011,1,151,12,5,0.01,0,'0'),(961305,950011,1,151,8,6,0.01,0,'0'),
 (961306,950011,1,151,10,6,0.01,0,'0'),(961307,950011,1,151,6,7,0.01,0,'0'),
 (961308,950011,1,151,12,7,0.01,0,'0'),(961309,950011,1,151,8,8,0.01,0,'0'),
 (961310,950011,1,151,10,8,0.01,0,'0'),(961311,950011,1,151,6,9,0.01,0,'0'),
 (961312,950011,1,151,9,9,0.01,0,'0'),(961313,950011,1,151,12,9,0.01,0,'0'),
 (961314,950011,1,151,7,10,0.01,0,'0'),(961315,950011,1,151,11,10,0.01,0,'0'),
 (961350,950013,1,151,4,4,0,0,'0'),(961351,950014,1,151,14,4,0,0,'0'),
 (961352,950015,1,151,4,10,0,0,'0'),(961353,950016,1,151,14,10,0,0,'0'),
 (961354,950017,1,151,2,5,0,0,'30'),(961355,950012,1,151,16,7,0,0,'0');

insert into furniture_items(id,definition_id,owner_player_id,room_id,x,y,z,rotation,extra_data)
overriding system value values
 (961400,950019,1,152,3,7,0,6,'0'),(961401,950020,1,152,17,7,0,2,'0'),
 (961402,950018,1,152,10,7,0,0,'0'),(961403,950021,1,152,4,11,0,0,'0'),
 (961404,950022,1,152,17,11,0,0,'0'),(961405,950023,1,152,9,11,0,0,'0'),
 (961406,950024,1,152,11,11,0,0,'30');

insert into furniture_items(id,definition_id,owner_player_id,room_id,x,y,z,rotation,extra_data)
overriding system value values
 (961500,950025,1,153,3,3,0,0,'0'),(961501,950025,1,153,5,3,0,0,'0'),
 (961502,950025,1,153,3,5,0,0,'0'),(961503,950025,1,153,5,5,0,0,'0'),
 (961504,950026,1,153,8,4,0,0,'0'),
 (961510,950027,1,153,10,3,0,0,'0'),(961511,950027,1,153,12,3,0,0,'0'),
 (961512,950027,1,153,10,5,0,0,'0'),(961513,950027,1,153,12,5,0,0,'0'),
 (961514,950030,1,153,15,4,0,0,'0'),
 (961520,950028,1,153,5,9,0,0,'0'),(961521,950028,1,153,7,9,0,0,'0'),
 (961522,950028,1,153,5,11,0,0,'0'),(961523,950028,1,153,7,11,0,0,'0'),
 (961524,950029,1,153,10,10,0,0,'0');

select setval(pg_get_serial_sequence('furniture_definitions','id'),greatest((select max(id) from furniture_definitions),1));
select setval(pg_get_serial_sequence('furniture_items','id'),greatest((select max(id) from furniture_items),1));
--rollback delete from room_wired_settings where item_id between 961000 and 961599; delete from furniture_items where id between 961000 and 961599; delete from furniture_definitions where id=950030;
