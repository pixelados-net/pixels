--liquibase formatted sql

--changeset pixels:furniture-seed-room-games-0038 context:development
insert into furniture_definitions(id,sprite_id,name,public_name,kind,width,length,stack_height,allow_stack,allow_walk,allow_sit,allow_lay,allow_inventory_stack,interaction_type,interaction_modes_count,multiheight,custom_params,metadata)
overriding system value values
 (900009,3681,'wf_act_reset_highscore','WIRED: Reset Highscore','floor',1,1,0.65,true,true,false,false,true,'wf_act_reset_highscore',4,'','','{"source":"pixels-games","wired_support":"client-editor"}'),
 (900010,3681,'wf_act_progress_achievement','WIRED: Progress Achievement','floor',1,1,0.65,true,true,false,false,true,'wf_act_progress_achievement',4,'','','{"source":"pixels-games","wired_support":"client-editor"}'),
 (900011,3681,'wf_act_progress_quest','WIRED: Progress Quest','floor',1,1,0.65,true,true,false,false,true,'wf_act_progress_quest',4,'','','{"source":"pixels-games","wired_support":"client-editor"}'),
 (900012,3681,'wf_act_start_quest','WIRED: Start Quest','floor',1,1,0.65,true,true,false,false,true,'wf_act_start_quest',4,'','','{"source":"pixels-games","wired_support":"client-editor"}'),
 (950001,3633,'bb_patch1','Banzai Arena Tile','floor',1,1,0.10,true,true,false,false,true,'battlebanzai_tile',12,'','','{"game":"banzai"}'),
 (950002,3628,'bb_gate_r','Red Banzai Gate','floor',1,1,0.10,true,true,false,false,true,'battlebanzai_gate_r',2,'','','{"game":"banzai"}'),
 (950003,3621,'bb_gate_g','Green Banzai Gate','floor',1,1,0.10,true,true,false,false,true,'battlebanzai_gate_g',2,'','','{"game":"banzai"}'),
 (950004,3643,'bb_gate_b','Blue Banzai Gate','floor',1,1,0.10,true,true,false,false,true,'battlebanzai_gate_b',2,'','','{"game":"banzai"}'),
 (950005,3635,'bb_gate_y','Yellow Banzai Gate','floor',1,1,0.10,true,true,false,false,true,'battlebanzai_gate_y',2,'','','{"game":"banzai"}'),
 (950006,3645,'bb_counter','Banzai Timer','floor',1,4,0.10,true,false,false,false,true,'game_timer',1,'','30,60,120,180,300,600','{"game":"banzai"}'),
 (950007,3637,'bb_apparatus','Banzai Sphere','floor',1,1,0.00,true,true,false,false,true,'battlebanzai_sphere',8,'','','{"game":"banzai"}'),
 (950008,3641,'bb_puck','Banzai Puck','floor',1,1,0.00,true,true,false,false,true,'battlebanzai_puck',3,'','','{"game":"banzai"}'),
 (950009,3642,'bb_rnd_tele','Banzai Random Teleport','floor',1,1,0.10,true,true,false,false,true,'battlebanzai_random_teleport',2,'','','{"game":"banzai"}'),
 (950010,3757,'es_tile','Freeze Tile','floor',1,1,0.00,true,true,false,false,true,'freeze_tile',20,'','','{"game":"freeze"}'),
 (950011,3758,'es_box','Freeze Ice Block','floor',1,1,0.50,true,false,false,false,true,'freeze_block',20,'','','{"game":"freeze"}'),
 (950012,3759,'es_exit','Freeze Exit','floor',1,1,0.10,true,true,false,false,true,'freeze_exit',2,'','','{"game":"freeze"}'),
 (950013,3760,'es_gate_r','Red Freeze Gate','floor',1,1,0.01,true,true,false,false,true,'freeze_gate_r',2,'','','{"game":"freeze"}'),
 (950014,3761,'es_gate_g','Green Freeze Gate','floor',1,1,0.01,true,true,false,false,true,'freeze_gate_g',2,'','','{"game":"freeze"}'),
 (950015,3762,'es_gate_b','Blue Freeze Gate','floor',1,1,0.01,true,true,false,false,true,'freeze_gate_b',2,'','','{"game":"freeze"}'),
 (950016,3763,'es_gate_y','Yellow Freeze Gate','floor',1,1,0.01,true,true,false,false,true,'freeze_gate_y',2,'','','{"game":"freeze"}'),
 (950017,3764,'es_counter','Freeze Timer','floor',1,3,1.00,true,false,false,false,true,'game_timer',1,'','30,60,120,180,300,600','{"game":"freeze"}'),
 (950018,3508,'fball_ball','Game Ball','floor',1,1,0.01,true,true,false,false,true,'football',2,'','','{"game":"football"}'),
 (950019,3514,'fball_goal_r','Red Football Goal','floor',3,1,0.01,true,true,false,false,true,'football_goal_red',2,'','','{"game":"football"}'),
 (950020,3515,'fball_goal_b','Blue Football Goal','floor',3,1,0.01,true,true,false,false,true,'football_goal_blue',2,'','','{"game":"football"}'),
 (950021,3522,'fball_score_r','Red Football Scoreboard','floor',1,1,1.00,true,false,false,false,true,'football_counter_red',100,'','','{"game":"football"}'),
 (950022,3496,'fball_score_b','Blue Football Scoreboard','floor',1,1,1.00,true,false,false,false,true,'football_counter_blue',100,'','','{"game":"football"}'),
 (950023,3516,'fball_gate','Football Dressing Gate','floor',1,1,0.00,true,true,false,false,true,'football_gate',2,'','ch-255-66.ca-1808-66.cc-3039-66.cp-3035-66.lg-275-66.wa-2001-66.sh-295-66','{"game":"football"}'),
 (950024,3525,'fball_counter','Football Timer','floor',2,1,1.00,true,false,false,false,true,'game_timer',1,'','30,60,120,180,300,600','{"game":"football"}'),
 (950025,3736,'es_skating_ice','IceTag Field','floor',2,2,0.01,true,true,false,false,true,'icetag_field',2,'','','{"game":"tag"}'),
 (950026,3741,'es_tagging','IceTag Pole','floor',1,1,3.00,true,false,false,false,true,'icetag_pole',2,'','','{"game":"tag"}'),
 (950027,3792,'val11_floor','Rollerskate Field','floor',2,2,0.10,true,true,false,false,true,'rollerskate_field',3,'','','{"game":"tag"}'),
 (950028,3909,'easter11_grasspatch','Bunnyrun Field','floor',2,2,0.00,true,true,false,false,true,'bunnyrun_field',2,'','','{"game":"tag"}'),
 (950029,3914,'easter11_tag','Bunnyrun Pole','floor',1,1,1.00,true,false,false,false,true,'bunnyrun_pole',2,'','','{"game":"tag"}')
on conflict(id) do update set sprite_id=excluded.sprite_id,name=excluded.name,public_name=excluded.public_name,
 width=excluded.width,length=excluded.length,stack_height=excluded.stack_height,allow_stack=excluded.allow_stack,
 allow_walk=excluded.allow_walk,interaction_type=excluded.interaction_type,interaction_modes_count=excluded.interaction_modes_count,
 custom_params=excluded.custom_params,metadata=excluded.metadata,deleted_at=null,updated_at=now();

insert into furniture_items(id,definition_id,owner_player_id,room_id,x,y,z,rotation,extra_data)
overriding system value
select 960000 + (y-3)*6 + (x-5),950001,1,150,x,y,0,0,'0'
from generate_series(5,10) x cross join generate_series(3,8) y
on conflict(id) do update set definition_id=excluded.definition_id,owner_player_id=1,room_id=150,x=excluded.x,y=excluded.y,z=0,rotation=0,extra_data='0',deleted_at=null;

insert into furniture_items(id,definition_id,owner_player_id,room_id,x,y,z,rotation,extra_data)
overriding system value values
 (423140,3675,1,113,10,3,0,0,'0'),(423141,900009,1,113,10,3,0.65,0,'0'),
 (425080,3675,1,115,4,3,0,0,'0'),(425081,900010,1,115,4,3,0.65,0,'0'),
 (425090,3675,1,115,5,3,0,0,'0'),(425091,900011,1,115,5,3,0.65,0,'0'),
 (425100,3675,1,115,6,3,0,0,'0'),(425101,900012,1,115,6,3,0.65,0,'0'),
 (960100,950002,1,150,4,3,0,0,'0'),(960101,950003,1,150,11,3,0,0,'0'),
 (960102,950004,1,150,4,8,0,0,'0'),(960103,950005,1,150,11,8,0,0,'0'),
 (960104,950006,1,150,4,10,0,0,'30'),(960105,950007,1,150,11,10,0,0,'0'),
 (960106,950008,1,150,7,6,0,0,'0'),(960107,950009,1,150,5,10,0,0,'0'),(960108,950009,1,150,10,10,0,0,'0'),
 (960109,900010,1,150,5,12,0,0,'0'),(960110,900011,1,150,6,12,0,0,'0'),(960111,900012,1,150,7,12,0,0,'0'),
 (960200,950013,1,151,4,3,0,0,'0'),(960201,950014,1,151,11,3,0,0,'0'),
 (960202,950015,1,151,4,8,0,0,'0'),(960203,950016,1,151,11,8,0,0,'0'),
 (960204,950017,1,151,4,10,0,0,'30'),(960205,950012,1,151,10,10,0,0,'0'),
 (960300,950019,1,152,4,6,0,6,'0'),(960301,950020,1,152,9,6,0,2,'0'),
 (960302,950018,1,152,7,6,0,0,'0'),(960303,950021,1,152,5,10,0,0,'0'),
 (960304,950022,1,152,9,10,0,0,'0'),(960305,950023,1,152,6,10,0,0,'0'),(960306,950024,1,152,7,10,0,0,'30'),
 (960400,950025,1,153,5,4,0,0,'0'),(960401,950025,1,153,7,4,0,0,'0'),
 (960402,950026,1,153,9,4,0,0,'0'),(960403,950028,1,153,5,8,0,0,'0'),
 (960404,950028,1,153,7,8,0,0,'0'),(960405,950029,1,153,9,8,0,0,'0')
on conflict(id) do update set definition_id=excluded.definition_id,owner_player_id=1,room_id=excluded.room_id,x=excluded.x,y=excluded.y,z=excluded.z,rotation=excluded.rotation,extra_data=excluded.extra_data,deleted_at=null;

insert into room_wired_settings(item_id,int_params,string_param,selection_mode,delay_pulses) values
 (423140,'[]','resethighscore',0,0),(423141,'[]','',1,0),
 (425080,'[]','achievementme',0,0),(425081,'[]','BattleBallPlayer',0,0),
 (425090,'[]','progressquest',0,0),(425091,'[]','900001',0,0),
 (425100,'[]','startquest',0,0),(425101,'[]','900001',0,0),
 (960109,'[]','BattleBallPlayer',0,0),(960110,'[]','900001',0,0),(960111,'[]','900001',0,0)
on conflict(item_id) do update set int_params=excluded.int_params,string_param=excluded.string_param,selection_mode=0,delay_pulses=0,updated_at=now();

insert into room_wired_selected_items(wired_item_id,selected_item_id,ordinal,snapshot_state,snapshot_x,snapshot_y,snapshot_z,snapshot_rotation)
values (423141,423210,0,null,null,null,null,null)
on conflict(wired_item_id,selected_item_id) do update set ordinal=excluded.ordinal,snapshot_state=excluded.snapshot_state,
 snapshot_x=excluded.snapshot_x,snapshot_y=excluded.snapshot_y,snapshot_z=excluded.snapshot_z,snapshot_rotation=excluded.snapshot_rotation;

insert into furniture_items(id,definition_id,owner_player_id,room_id,x,y,z,rotation,extra_data)
overriding system value
select 960220 + (y-4)*4 + (x-6),case when (x+y)%3=0 then 950011 else 950010 end,1,151,x,y,0,0,'0'
from generate_series(6,9) x cross join generate_series(4,7) y
on conflict(id) do update set definition_id=excluded.definition_id,owner_player_id=1,room_id=151,x=excluded.x,y=excluded.y,z=0,rotation=0,extra_data='0',deleted_at=null;

select setval(pg_get_serial_sequence('furniture_definitions','id'),greatest((select max(id) from furniture_definitions),1));
select setval(pg_get_serial_sequence('furniture_items','id'),greatest((select max(id) from furniture_items),1));
--rollback delete from room_wired_selected_items where wired_item_id=423141 and selected_item_id=423210; delete from furniture_items where id in (423140,423141,425080,425081,425090,425091,425100,425101) or id between 960000 and 960499; delete from furniture_definitions where id between 900009 and 900012 or id between 950001 and 950029;
