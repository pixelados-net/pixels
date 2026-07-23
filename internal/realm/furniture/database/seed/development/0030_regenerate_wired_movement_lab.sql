--liquibase formatted sql

--changeset pixels:pixels-furniture-seed-development-0030-regenerate-wired-movement-lab context:development
-- Rebuild room 112 and add one isolated, repeatable furniture-to-unit collision fixture.
delete from furniture_items where room_id=112;

insert into furniture_items(id,definition_id,owner_player_id,room_id,x,y,z,rotation,extra_data)
overriding system value
values
    (422000,3683,1,112,4,1,0,0,'0'),(422001,3681,1,112,4,1,0.65,0,'0'),
    (422010,3703,1,112,5,1,0,0,'0'),(422011,3681,1,112,5,1,0.65,0,'0'),
    (422020,3678,1,112,6,1,0,0,'0'),(422021,3681,1,112,6,1,0.65,0,'0'),
    (422030,5050,1,112,7,1,0,0,'0'),(422031,3681,1,112,7,1,0.65,0,'0'),
    (422040,3675,1,112,8,1,0,0,'0'),(422041,3663,1,112,8,1,0.65,0,'0'),
    (422050,3675,1,112,9,1,0,0,'0'),(422051,5055,1,112,9,1,0.65,0,'0'),
    (422060,3675,1,112,10,1,0,0,'0'),(422061,5061,1,112,10,1,0.65,0,'0'),
    (422070,3675,1,112,11,1,0,0,'0'),(422071,3700,1,112,11,1,0.65,0,'0'),
    (422080,3675,1,112,4,3,0,0,'0'),(422081,5450,1,112,4,3,0.65,0,'0'),
    (422090,3675,1,112,5,3,0,0,'0'),(422091,5048,1,112,5,3,0.65,0,'0'),
    (422100,3675,1,112,6,3,0,0,'0'),(422101,3674,1,112,6,3,0.65,0,'0'),
    (422110,3675,1,112,7,3,0,0,'0'),(422111,5451,1,112,7,3,0.65,0,'0'),
    (422120,3675,1,112,8,3,0,0,'0'),(422121,3685,1,112,8,3,0.65,0,'0'),
    (422130,3675,1,112,9,3,0,0,'0'),(422131,900008,1,112,9,3,0.65,0,'0'),
    (422132,3663,1,112,9,3,1.30,0,'0'),(422133,3681,1,112,9,3,1.95,0,'0'),
    (422140,3668,1,112,10,3,0,0,'0'),(422141,3681,1,112,10,3,0.65,0,'0'),
    (422150,3675,1,112,11,3,0,0,'0'),(422151,5048,1,112,11,3,0.65,0,'0'),
    (429200,28,1,112,4,12,0,0,'0'),(429201,27,1,112,7,10,0,0,'0'),
    (429202,30100,1,112,10,10,0,0,'0'),(429203,29,1,112,10,8,0,0,'0')
on conflict(id) do update set definition_id=excluded.definition_id,owner_player_id=excluded.owner_player_id,
    room_id=excluded.room_id,x=excluded.x,y=excluded.y,z=excluded.z,rotation=excluded.rotation,
    extra_data=excluded.extra_data,updated_at=now();

insert into room_wired_settings(item_id,int_params,string_param,selection_mode,delay_pulses)
values
    (422000,'[]','',0,0),(422001,'[]','Movement lab ready. Stand at (11,8), next to the Floor Switch at (10,8), say collision twice, then test rotate, chase, flee, snapshot, moveto, direction, teleport, randomstate, togglestate and validmove.',0,0),
    (422010,'[]','',1,0),(422011,'[]','PASS walk-on fired.',0,0),
    (422020,'[]','',1,0),(422021,'[]','PASS walk-off fired.',0,0),
    (422030,'[]','',0,0),(422031,'[]','PASS collision fired.',0,0),
    (422040,'[]','rotate',0,0),(422041,'[0]','',1,0),
    (422050,'[]','chase',0,0),(422051,'[0]','',1,0),
    (422060,'[]','flee',0,0),(422061,'[0]','',1,0),
    (422070,'[]','snapshot',0,0),(422071,'[1,1,1]','',1,0),
    (422080,'[]','moveto',0,0),(422081,'[2]','',1,0),
    (422090,'[]','direction',0,0),(422091,'[4]','',1,0),
    (422100,'[]','teleport',0,0),(422101,'[]','',1,0),
    (422110,'[]','randomstate',0,0),(422111,'[]','',1,0),
    (422120,'[]','togglestate',0,0),(422121,'[]','',1,0),
    (422130,'[]','validmove',0,0),(422131,'[]','',0,0),
    (422132,'[2]','',1,0),(422133,'[]','PASS valid move simulated and applied.',0,0),
    (422140,'[]','',1,0),(422141,'[]','PASS selected furniture state changed.',0,0),
    (422150,'[]','collision',0,0),(422151,'[2]','',1,0)
on conflict(item_id) do update set int_params=excluded.int_params,string_param=excluded.string_param,
    selection_mode=excluded.selection_mode,delay_pulses=excluded.delay_pulses,updated_at=now();

insert into room_wired_selected_items(wired_item_id,selected_item_id,ordinal,snapshot_state,snapshot_x,snapshot_y,snapshot_z,snapshot_rotation)
values
    (422010,429200,0,null,null,null,null,null),(422020,429200,0,null,null,null,null,null),
    (422041,429201,0,null,null,null,null,null),(422051,429201,0,null,null,null,null,null),
    (422061,429201,0,null,null,null,null,null),(422071,429201,0,'0',7,10,0,0),
    (422081,429201,0,null,null,null,null,null),(422091,429201,0,null,null,null,null,null),
    (422101,429202,0,null,null,null,null,null),(422111,429201,0,null,null,null,null,null),
    (422121,429201,0,null,null,null,null,null),(422132,429201,0,null,null,null,null,null),
    (422140,429201,0,null,null,null,null,null),(422151,429203,0,null,null,null,null,null)
on conflict(wired_item_id,selected_item_id) do update set ordinal=excluded.ordinal,
    snapshot_state=excluded.snapshot_state,snapshot_x=excluded.snapshot_x,
    snapshot_y=excluded.snapshot_y,snapshot_z=excluded.snapshot_z,
    snapshot_rotation=excluded.snapshot_rotation;

select setval(pg_get_serial_sequence('furniture_items','id'),greatest((select max(id) from furniture_items),1));

--rollback delete from furniture_items where room_id=112;
