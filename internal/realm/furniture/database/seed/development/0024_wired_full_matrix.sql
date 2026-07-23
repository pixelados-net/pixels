--liquibase formatted sql

--changeset pixels:pixels-furniture-seed-development-0024-wired-full-matrix context:development
-- Completes the six-room QA matrix so every canonical behavior and the ms3 extension is placed and configured.
insert into furniture_items(id,definition_id,owner_player_id,room_id,x,y,z,rotation,extra_data)
overriding system value
values
    (417000,5056,1,110,4,4,0,0,'0'),(417001,5861,1,110,5,4,0,0,'0'),
    (417002,3692,1,110,7,4,0,0,'0'),(417003,3857,1,110,8,4,0,0,'0'),
    (417004,7854,1,110,9,4,0,0,'0'),(417005,3695,1,110,10,4,0,0,'0'),
    (417006,5440,1,110,11,4,0,0,'0'),(417007,5441,1,110,4,6,0,0,'0'),
    (417008,5448,1,110,5,6,0,0,'0'),(417009,5439,1,110,6,6,0,0,'0'),
    (417010,5452,1,110,7,6,0,0,'0'),(417011,5449,1,110,9,6,0,0,'0'),
    (417012,5438,1,110,10,6,0,0,'0'),(417013,5443,1,110,11,6,0,0,'0'),
    (417014,5446,1,110,4,8,0,0,'0'),(417015,5444,1,110,5,8,0,0,'0'),
    (417016,5447,1,110,6,8,0,0,'0'),(417017,3682,1,110,7,8,0,0,'0'),
    (417018,3665,1,110,8,8,0,0,'0'),(417019,3694,1,110,9,8,0,0,'0'),
    (417020,5445,1,110,10,8,0,0,'0'),(417021,5095,1,110,11,8,0,0,'0'),
    (417100,3679,1,111,4,4,0,0,'0'),(417101,5442,1,111,5,4,0,0,'0'),
    (417102,7856,1,111,6,4,0,0,'0'),(417103,7850,1,111,7,4,0,0,'0'),
    (417104,3680,1,111,8,4,0,0,'0'),(417105,5042,1,111,9,4,0,0,'0'),
    (417106,3668,1,111,10,4,0,0,'0'),(417107,3678,1,111,11,4,0,0,'0'),
    (417150,3681,1,111,4,4,0.65,0,'0'),(417151,3681,1,111,5,4,0.65,0,'0'),
    (417152,3681,1,111,6,4,0.65,0,'0'),(417153,3681,1,111,7,4,0.65,0,'0'),
    (417154,3681,1,111,8,4,0.65,0,'0'),(417155,3681,1,111,9,4,0.65,0,'0'),
    (417156,3681,1,111,10,4,0.65,0,'0'),(417157,3681,1,111,11,4,0.65,0,'0'),
    (417158,3675,1,111,4,6,0,0,'0'),(417108,4834,1,111,4,6,0.65,0,'0'),
    (417200,5055,1,112,4,8,0.65,0,'0'),(417201,5061,1,112,5,8,0.65,0,'0'),
    (417202,3700,1,112,6,8,0.65,0,'0'),(417203,5450,1,112,7,8,0.65,0,'0'),
    (417204,5048,1,112,8,8,0.65,0,'0'),(417205,3674,1,112,9,8,0.65,0,'0'),
    (417206,5451,1,112,10,8,0.65,0,'0'),
    (417250,3675,1,112,4,8,0,0,'0'),(417251,3675,1,112,5,8,0,0,'0'),
    (417252,3675,1,112,6,8,0,0,'0'),(417253,3675,1,112,7,8,0,0,'0'),
    (417254,3675,1,112,8,8,0,0,'0'),(417255,3675,1,112,9,8,0,0,'0'),
    (417256,3675,1,112,10,8,0,0,'0'),
    (417300,3697,1,113,4,4,0.65,0,'0'),(417301,5043,1,113,4,6,0.65,0,'0'),
    (417350,3675,1,113,4,4,0,0,'0'),(417351,3675,1,113,4,6,0,0,'0'),
    (417400,7852,1,114,4,10,0.65,0,'0'),(417401,7855,1,114,5,10,0.65,0,'0'),
    (417402,7849,1,114,6,10,0.65,0,'0'),(417450,3675,1,114,4,10,0,0,'0'),
    (417451,3675,1,114,5,10,0,0,'0'),(417452,3675,1,114,6,10,0,0,'0'),
    (419901,2,1,111,11,12,0,0,'0'),(419903,2,1,113,11,12,0,0,'0')
on conflict(id) do update set definition_id=excluded.definition_id,owner_player_id=excluded.owner_player_id,
    room_id=excluded.room_id,x=excluded.x,y=excluded.y,z=excluded.z,rotation=excluded.rotation,extra_data=excluded.extra_data;

insert into room_wired_settings(item_id,int_params,string_param,selection_mode,delay_pulses)
values
    (417000,'[1]','',0,0),(417001,'[0,2147483647]','',0,0),(417002,'[]','',1,0),
    (417003,'[1]','',1,0),(417004,'[1]','',0,0),(417005,'[1,1,1]','',1,0),
    (417006,'[1]','',1,0),(417007,'[]','',1,0),(417008,'[]','',0,0),
    (417009,'[1]','',0,0),(417010,'[1,1,1]','',1,0),(417011,'[]','',1,0),
    (417012,'[]','',1,0),(417013,'[0,99]','',0,0),(417014,'[]','ACH_WIREDQA1',0,0),
    (417015,'[4]','',0,0),(417016,'[]','',1,0),(417017,'[2]','',0,0),
    (417018,'[2]','',0,0),(417019,'[]','',1,0),(417020,'[0,99]','',0,0),
    (417021,'[4]','',0,0),
    (417100,'[2]','',0,0),(417101,'[2]','',0,0),(417102,'[]','WiredRunner',0,0),
    (417103,'[]','WiredRunner',1,0),(417104,'[]','',0,0),(417105,'[1]','',0,0),
    (417106,'[]','',1,0),(417107,'[]','',1,0),
    (417150,'[]','At-time fired.',0,0),(417151,'[]','Long at-time fired.',0,0),
    (417152,'[]','Bot reached avatar.',0,0),(417153,'[]','Bot reached target.',0,0),
    (417154,'[]','Game ended.',0,0),(417155,'[]','Long period fired.',0,0),
    (417156,'[]','Selected furniture changed.',0,0),(417157,'[]','Walk-off fired.',0,0),
    (417158,'[]','callstacks',0,0),(417108,'[]','',1,0),
    (417200,'[0]','',1,0),(417201,'[0]','',1,0),(417202,'[1,1,1]','',1,0),
    (417203,'[2]','',1,0),(417204,'[2]','',1,0),(417205,'[]','',1,0),
    (417206,'[]','',1,0),(417250,'[]','chase',0,0),(417251,'[]','flee',0,0),
    (417252,'[]','snapshot',0,0),(417253,'[]','moveto',0,0),(417254,'[]','direction',0,0),
    (417255,'[]','teleport',0,0),(417256,'[]','randomstate',0,0),
    (417300,'[1,10]','',0,0),(417301,'[1,1]','',0,0),
    (417350,'[]','score',0,0),(417351,'[]','teamscore',0,0),
    (417400,'[1]','WiredGuide',0,0),(417401,'[]',E'WiredGuide\tThis is a private WIRED bot message.',0,0),
    (417402,'[]','WiredRunner',1,0),(417450,'[]','botitem',0,0),
    (417451,'[]','botwhisper',0,0),(417452,'[]','botteleport',0,0)
on conflict(item_id) do update set int_params=excluded.int_params,string_param=excluded.string_param,
    selection_mode=excluded.selection_mode,delay_pulses=excluded.delay_pulses,updated_at=now();

insert into room_wired_selected_items(wired_item_id,selected_item_id,ordinal,snapshot_state,snapshot_x,snapshot_y,snapshot_z,snapshot_rotation)
values
    (417002,410004,0,null,null,null,null,null),(417003,410004,0,null,null,null,null,null),
    (417005,410004,0,'0',8,6,0,0),(417006,410004,0,null,null,null,null,null),
    (417007,410004,0,null,null,null,null,null),(417010,410004,0,'0',8,6,0,0),
    (417011,410004,0,null,null,null,null,null),(417012,410004,0,null,null,null,null,null),
    (417016,410004,0,null,null,null,null,null),(417019,410004,0,null,null,null,null,null),
    (417103,419901,0,null,null,null,null,null),(417106,419901,0,null,null,null,null,null),
    (417107,419901,0,null,null,null,null,null),(417108,411001,0,null,null,null,null,null),
    (417200,412004,0,null,null,null,null,null),(417201,412004,0,null,null,null,null,null),
    (417202,412004,0,'0',8,6,0,0),(417203,412004,0,null,null,null,null,null),
    (417204,412004,0,null,null,null,null,null),(417205,412004,0,null,null,null,null,null),
    (417206,412004,0,null,null,null,null,null),(417402,414008,0,null,null,null,null,null)
on conflict(wired_item_id,selected_item_id) do update set ordinal=excluded.ordinal,snapshot_state=excluded.snapshot_state,
    snapshot_x=excluded.snapshot_x,snapshot_y=excluded.snapshot_y,snapshot_z=excluded.snapshot_z,snapshot_rotation=excluded.snapshot_rotation;

select setval(pg_get_serial_sequence('furniture_items','id'),greatest((select max(id) from furniture_items),1));
--rollback delete from furniture_items where id between 417000 and 419999;
