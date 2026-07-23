--liquibase formatted sql

--changeset pixels:pixels-furniture-seed-development-0031-wired-lab-extras context:development
--validCheckSum:ANY
-- Room 116: cases the six original labs never covered — roller-driven walk-on/off,
-- multi-tile carpet footprints, username-filtered room entry and per-minute rewards.
insert into furniture_items(id,definition_id,owner_player_id,room_id,x,y,z,rotation,extra_data)
overriding system value
values
    -- Roller feeding the pressure plate: roller at (4,5) rotated east pushes onto the plate at (5,5).
    (426100,28,1,116,5,5,0,0,'0'),
    (426101,7,1,116,4,5,0,2,'0'),
    -- Multi-tile target: 3x3 walkable carpet, single selected furni with a 9-tile footprint.
    (426110,30087,1,116,7,4,0,0,'0'),

    -- Stack: walk-on selecting the plate (roller-driven entry must fire it once, after the roll delay).
    (426000,3703,1,116,4,1,0,0,'0'),(426001,3681,1,116,4,1,0.65,0,'0'),
    -- Stack: walk-off selecting the plate (rolling off or stepping off fires it once).
    (426010,3678,1,116,5,1,0,0,'0'),(426011,3681,1,116,5,1,0.65,0,'0'),
    -- Stack: walk-on selecting the carpet (crossing between its own tiles must NOT re-fire).
    (426020,3703,1,116,6,1,0,0,'0'),(426021,3681,1,116,6,1,0.65,0,'0'),
    -- Stack: walk-off selecting the carpet.
    (426030,3678,1,116,7,1,0,0,'0'),(426031,3681,1,116,7,1,0.65,0,'0'),
    -- Stack: enter-room filtered to juno only.
    (426040,3683,1,116,8,1,0,0,'0'),(426041,3681,1,116,8,1,0.65,0,'0'),
    -- Stack: say 'minutereward' claims a per-minute unique credits reward.
    (426050,3675,1,116,9,1,0,0,'0'),(426051,4505,1,116,9,1,0.65,0,'0'),
    -- Stack: unfiltered entry guide.
    (426060,3683,1,116,10,1,0,0,'0'),(426061,3681,1,116,10,1,0.65,0,'0')
on conflict(id) do update set definition_id=excluded.definition_id,owner_player_id=excluded.owner_player_id,
    room_id=excluded.room_id,x=excluded.x,y=excluded.y,z=excluded.z,rotation=excluded.rotation,extra_data=excluded.extra_data;

insert into room_wired_settings(item_id,int_params,string_param,selection_mode,delay_pulses)
values
    (426000,'[]','',0,0),(426001,'[]','PASS roller walk-on fired.',0,0),
    (426010,'[]','',0,0),(426011,'[]','PASS roller walk-off fired.',0,0),
    (426020,'[]','',0,0),(426021,'[]','PASS carpet entered.',0,0),
    (426030,'[]','',0,0),(426031,'[]','PASS carpet left.',0,0),
    (426040,'[]','juno',0,0),(426041,'[]','PASS filtered entry for juno.',0,0),
    (426050,'[]','minutereward',0,0),(426051,'[3,1,0,1]','',0,0),
    (426060,'[]','',0,0),(426061,'[]','Extras lab: ride the roller onto the plate, cross the carpet, re-enter as juno, say minutereward.',0,0)
on conflict(item_id) do update set int_params=excluded.int_params,string_param=excluded.string_param,
    selection_mode=excluded.selection_mode,delay_pulses=excluded.delay_pulses,updated_at=now();

insert into room_wired_selected_items(wired_item_id,selected_item_id,ordinal,snapshot_state,snapshot_x,snapshot_y,snapshot_z,snapshot_rotation)
values
    (426000,426100,0,null,null,null,null,null),
    (426010,426100,0,null,null,null,null,null),
    (426020,426110,0,null,null,null,null,null),
    (426030,426110,0,null,null,null,null,null)
on conflict(wired_item_id,selected_item_id) do update set ordinal=excluded.ordinal,snapshot_state=excluded.snapshot_state,
    snapshot_x=excluded.snapshot_x,snapshot_y=excluded.snapshot_y,snapshot_z=excluded.snapshot_z,snapshot_rotation=excluded.snapshot_rotation;

insert into room_wired_rewards(wired_item_id,ordinal,kind,reference,amount,weight,stock)
values
    (426051,0,'credits','credits',5,1,null)
on conflict(wired_item_id,ordinal) do update set kind=excluded.kind,reference=excluded.reference,
    amount=excluded.amount,weight=excluded.weight,stock=excluded.stock;

select setval(pg_get_serial_sequence('furniture_items','id'),greatest((select max(id) from furniture_items),1));

--rollback delete from room_wired_rewards where wired_item_id between 426000 and 426999;
--rollback delete from room_wired_selected_items where wired_item_id between 426000 and 426999;
--rollback delete from room_wired_settings where item_id between 426000 and 426999;
--rollback delete from furniture_items where id between 426000 and 426999;
