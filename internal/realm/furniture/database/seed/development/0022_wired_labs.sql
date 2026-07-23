--liquibase formatted sql

--changeset pixels:pixels-furniture-seed-development-0022-wired-labs context:development
-- Items sharing x/y form one deterministic stack; comments in room descriptions identify each lab.
insert into furniture_items(id,definition_id,owner_player_id,room_id,x,y,z,rotation,extra_data)
overriding system value
values
    (410000,3683,1,110,4,2,0,0,'0'),(410001,3681,1,110,4,2,0.65,0,'0'),
    (410002,3675,1,110,6,4,0,0,'0'),(410003,3685,1,110,6,4,0.65,0,'0'),(410004,2,1,110,8,6,0,0,'0'),
    (411000,3671,1,111,4,2,0,0,'0'),(411001,3681,1,111,4,2,0.65,0,'0'),(411002,3669,1,111,4,2,1.30,0,'0'),
    (411003,3681,1,111,4,2,1.95,0,'0'),(411004,3670,1,111,4,2,2.60,0,'0'),(411005,3691,1,111,7,5,0,0,'0'),
    (412000,3703,1,112,4,2,0,0,'0'),(412001,3663,1,112,4,2,0.65,0,'0'),(412002,5050,1,112,6,4,0,0,'0'),
    (412003,3681,1,112,6,4,0.65,0,'0'),(412004,1,1,112,8,6,0,0,'0'),
    (413000,3702,1,113,4,2,0,0,'0'),(413001,3681,1,113,4,2,0.65,0,'0'),
    (413002,3675,1,113,6,4,0,0,'0'),(413003,5062,1,113,6,4,0.65,0,'0'),
    (413004,3673,1,113,8,6,0,0,'0'),(413005,3681,1,113,8,6,0.65,0,'0'),
    (413006,5067,1,113,10,8,0,0,'1'),(413007,5068,1,113,11,8,0,0,'1'),
    (413008,5044,1,113,5,8,0,0,'{"state":"0","score_type":2,"clear_type":0,"entries":[]}'),
    (413009,5051,1,113,6,8,0,0,'{"state":"0","score_type":0,"clear_type":0,"entries":[]}'),
    (413010,5057,1,113,7,8,0,0,'{"state":"0","score_type":1,"clear_type":0,"entries":[]}'),
    (413011,5045,1,113,5,10,0,0,'{"state":"0","score_type":2,"clear_type":1,"entries":[]}'),
    (413012,5046,1,113,6,10,0,0,'{"state":"0","score_type":2,"clear_type":2,"entries":[]}'),
    (413013,5047,1,113,7,10,0,0,'{"state":"0","score_type":2,"clear_type":3,"entries":[]}'),
    (413014,5052,1,113,8,10,0,0,'{"state":"0","score_type":0,"clear_type":1,"entries":[]}'),
    (413015,5053,1,113,9,10,0,0,'{"state":"0","score_type":0,"clear_type":2,"entries":[]}'),
    (413016,5054,1,113,10,10,0,0,'{"state":"0","score_type":0,"clear_type":3,"entries":[]}'),
    (413017,5058,1,113,5,12,0,0,'{"state":"0","score_type":1,"clear_type":1,"entries":[]}'),
    (413018,5059,1,113,6,12,0,0,'{"state":"0","score_type":1,"clear_type":2,"entries":[]}'),
    (413019,5060,1,113,7,12,0,0,'{"state":"0","score_type":1,"clear_type":3,"entries":[]}'),
    (413020,3675,1,113,8,2,0,0,'0'),(413021,5062,1,113,8,2,0.65,0,'0'),
    (413022,3675,1,113,8,4,0,0,'0'),(413023,5062,1,113,8,4,0.65,0,'0'),
    (413024,3675,1,113,10,6,0,0,'0'),(413025,5062,1,113,10,6,0.65,0,'0'),
    (413026,3675,1,113,9,6,0,0,'0'),(413027,5062,1,113,9,6,0.65,0,'0'),
    (413028,3675,1,113,10,12,0,0,'0'),(413029,5049,1,113,10,12,0.65,0,'0'),
    (414000,3683,1,114,4,2,0,0,'0'),(414001,7857,1,114,4,2,0.65,0,'0'),
    (414002,3675,1,114,6,4,0,0,'0'),(414003,7853,1,114,6,4,0.65,0,'0'),
    (414004,3675,1,114,8,6,0,0,'0'),(414005,7851,1,114,8,6,0.65,0,'0'),
    (414006,3675,1,114,10,8,0,0,'0'),(414007,7848,1,114,10,8,0.65,0,'0'),(414008,2,1,114,11,8,0,0,'0'),
    (415000,3683,1,115,4,2,0,0,'0'),(415001,4505,1,115,4,2,0.65,0,'0'),
    (415002,3675,1,115,6,4,0,0,'0'),(415003,4947,1,115,6,4,0.65,0,'0'),
    (415004,3675,1,115,8,6,0,0,'0'),(415005,5854,1,115,8,6,0.65,0,'0'),(415006,2,1,115,10,8,0,0,'0')
on conflict(id) do update set definition_id=excluded.definition_id,owner_player_id=excluded.owner_player_id,
    room_id=excluded.room_id,x=excluded.x,y=excluded.y,z=excluded.z,rotation=excluded.rotation,extra_data=excluded.extra_data;

insert into room_wired_settings(item_id,int_params,string_param,selection_mode,delay_pulses)
values
    (410000,'[]','',0,0),(410001,'[]','Welcome to the avatar WIRED lab.',0,0),
    (410002,'[]','toggle',0,0),(410003,'[]','',1,0),
    (411000,'[2]','',0,0),(411001,'[]','Periodic branch A.',0,0),(411002,'[]','',0,0),
    (411003,'[]','Random branch B.',0,2),(411004,'[]','',0,0),(411005,'[]','',0,0),
    (412000,'[]','',1,0),(412001,'[0]','',1,0),(412002,'[]','',0,0),(412003,'[]','Collision detected.',0,0),
    (413000,'[]','',0,0),(413001,'[]','Game started.',0,0),(413002,'[]','team',0,0),(413003,'[1]','',0,0),
    (413004,'[10]','',0,0),(413005,'[]','Ten points reached.',0,0),
    (413020,'[]','red',0,0),(413021,'[1]','',0,0),(413022,'[]','green',0,0),(413023,'[2]','',0,0),
    (413024,'[]','blue',0,0),(413025,'[3]','',0,0),(413026,'[]','yellow',0,0),(413027,'[4]','',0,0),
    (413028,'[]','leave',0,0),(413029,'[]','',0,0),
    (414000,'[]','',0,0),(414001,'[]',E'WiredGuide\tWelcome to the bot lab.',0,0),
    (414002,'[]','movebot',0,0),(414003,'[]','WiredRunner',1,0),
    (414004,'[]','followbot',0,0),(414005,'[]','WiredRunner',0,0),
    (414006,'[]','clothesbot',0,0),(414007,'[]',E'WiredGuide\thd-180-1.ch-210-66.lg-270-82.sh-290-80',0,0),
    (415000,'[]','',0,0),(415001,'[0,0,0,1]','1,credits#10,10;0,WIRED_QA,2;1,furni#2,2;1,respect#1,1',0,0),
    (415002,'[]','kickme',0,0),(415003,'[]','',0,0),(415004,'[]','muteme',0,0),(415005,'[1]','',0,0)
on conflict(item_id) do update set int_params=excluded.int_params,string_param=excluded.string_param,
    selection_mode=excluded.selection_mode,delay_pulses=excluded.delay_pulses,updated_at=now();

insert into room_wired_selected_items(wired_item_id,selected_item_id,ordinal)
values (410003,410004,0),(412000,412004,0),(412001,412004,0),(414003,414008,0)
on conflict(wired_item_id,selected_item_id) do update set ordinal=excluded.ordinal;

insert into room_wired_rewards(wired_item_id,ordinal,kind,reference,amount,weight,stock)
values
    (415001,0,'credits','credits',10,10,null),(415001,1,'badge','WIRED_QA',1,2,null),
    (415001,2,'furniture','2',1,2,10),(415001,3,'respect','respect',1,1,null)
on conflict(wired_item_id,ordinal) do update set kind=excluded.kind,reference=excluded.reference,
    amount=excluded.amount,weight=excluded.weight,stock=excluded.stock;

insert into room_social_groups(room_id,group_id)
values (110,1)
on conflict(room_id) do update set group_id=excluded.group_id;

select setval(pg_get_serial_sequence('furniture_items','id'),greatest((select max(id) from furniture_items),1));
--rollback delete from furniture_items where id between 410000 and 415999;
