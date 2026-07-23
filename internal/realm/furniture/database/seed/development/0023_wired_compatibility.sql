--liquibase formatted sql

--changeset pixels:pixels-furniture-seed-development-0023-wired-compatibility context:development
-- These API/dev-only definitions reuse verified WIRED box sprites because Nitro has no distinct editor asset.
insert into furniture_definitions(id,sprite_id,name,public_name,kind,width,length,stack_height,allow_stack,allow_walk,allow_sit,allow_lay,allow_inventory_stack,interaction_type,interaction_modes_count,multiheight,custom_params,metadata)
overriding system value
values
    (900001,3681,'wf_act_alert','WIRED Compatibility: Alert','floor',1,1,0.65,true,true,false,false,true,'wf_act_alert',4,'','','{"source":"pixels-wired-compatibility","wired_support":"api-only"}'),
    (900002,3681,'wf_act_give_respect','WIRED Compatibility: Give Respect','floor',1,1,0.65,true,true,false,false,true,'wf_act_give_respect',4,'','','{"source":"pixels-wired-compatibility","wired_support":"api-only"}'),
    (900003,3681,'wf_act_give_handitem','WIRED Compatibility: Give Hand Item','floor',1,1,0.65,true,true,false,false,true,'wf_act_give_handitem',4,'','','{"source":"pixels-wired-compatibility","wired_support":"api-only"}'),
    (900004,3681,'wf_act_give_effect','WIRED Compatibility: Give Effect','floor',1,1,0.65,true,true,false,false,true,'wf_act_give_effect',4,'','','{"source":"pixels-wired-compatibility","wired_support":"api-only"}'),
    (900005,7850,'wf_trg_game_team_win','WIRED Compatibility: Team Wins','floor',1,1,0.65,true,true,false,false,true,'wf_trg_game_team_win',1,'','','{"source":"pixels-wired-compatibility","wired_support":"api-only"}'),
    (900006,7850,'wf_trg_game_team_lose','WIRED Compatibility: Team Loses','floor',1,1,0.65,true,true,false,false,true,'wf_trg_game_team_lose',1,'','','{"source":"pixels-wired-compatibility","wired_support":"api-only"}'),
    (900007,3669,'wf_xtra_or_eval','WIRED Add-on: OR Conditions','floor',1,1,0.65,true,true,false,false,true,'wf_xtra_or_eval',4,'','','{"source":"pixels-wired-compatibility","wired_support":"api-only"}'),
    (900008,3695,'wf_cnd_valid_moves','WIRED Compatibility: Valid Moves','floor',1,1,0.65,true,true,false,false,true,'wf_cnd_valid_moves',2,'','','{"source":"pixels-wired-compatibility","wired_support":"api-only"}')
on conflict(id) do update set interaction_type=excluded.interaction_type,metadata=excluded.metadata;

insert into furniture_items(id,definition_id,owner_player_id,room_id,x,y,z,rotation,extra_data)
overriding system value
values
    (416000,3683,1,110,10,2,0,0,'0'),(416001,5094,1,110,10,2,0.65,0,'0'),
    (416002,4281,1,110,10,2,1.30,0,'0'),(416003,900007,1,110,10,2,1.95,0,'0'),
    (416004,3681,1,110,10,2,2.60,0,'0'),
    (416100,3683,1,112,10,2,0,0,'0'),(416101,900008,1,112,10,2,0.65,0,'0'),
    (416102,3663,1,112,10,2,1.30,0,'0'),
    (416200,900005,1,113,10,2,0,0,'0'),(416201,3681,1,113,10,2,0.65,0,'0'),
    (416202,900006,1,113,10,4,0,0,'0'),(416203,3681,1,113,10,4,0.65,0,'0'),
    (416300,3675,1,115,10,2,0,0,'0'),(416301,900001,1,115,10,2,0.65,0,'0'),
    (416302,3675,1,115,10,4,0,0,'0'),(416303,900002,1,115,10,4,0.65,0,'0'),
    (416304,3675,1,115,10,6,0,0,'0'),(416305,900003,1,115,10,6,0.65,0,'0'),
    (416306,3675,1,115,6,8,0,0,'0'),(416307,900004,1,115,6,8,0.65,0,'0')
on conflict(id) do update set definition_id=excluded.definition_id,room_id=excluded.room_id,x=excluded.x,y=excluded.y,z=excluded.z;

insert into room_wired_settings(item_id,int_params,string_param,selection_mode,delay_pulses)
values
    (416000,'[]','',0,0),(416001,'[]','ACH_WIREDQA1',0,0),(416002,'[]','',0,0),
    (416003,'[]','',0,0),(416004,'[]','Badge OR social-group condition passed.',0,0),
    (416100,'[]','',0,0),(416101,'[]','',0,0),(416102,'[2]','',1,0),
    (416200,'[]','',0,0),(416201,'[]','Your team won.',0,0),
    (416202,'[]','',0,0),(416203,'[]','Your team lost.',0,0),
    (416300,'[]','alertme',0,0),(416301,'[]','Compatibility alert for %username%.',0,0),
    (416302,'[]','respectme',0,0),(416303,'[]','1',0,0),
    (416304,'[]','handme',0,0),(416305,'[]','1',0,0),
    (416306,'[]','effectme',0,0),(416307,'[]','4',0,0)
on conflict(item_id) do update set int_params=excluded.int_params,string_param=excluded.string_param,selection_mode=excluded.selection_mode,delay_pulses=excluded.delay_pulses;

insert into room_wired_selected_items(wired_item_id,selected_item_id,ordinal)
values (416102,412004,0)
on conflict(wired_item_id,selected_item_id) do nothing;

select setval(pg_get_serial_sequence('furniture_definitions','id'),greatest((select max(id) from furniture_definitions),1));
select setval(pg_get_serial_sequence('furniture_items','id'),greatest((select max(id) from furniture_items),1));
--rollback delete from furniture_items where id between 416000 and 416399; delete from furniture_definitions where id between 900001 and 900008;
