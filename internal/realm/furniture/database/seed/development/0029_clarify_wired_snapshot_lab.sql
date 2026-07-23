--liquibase formatted sql

--changeset pixels:pixels-furniture-seed-development-0029-clarify-wired-snapshot-lab context:development
-- Give the snapshot fixture a unique sprite and explain its exact target on room entry.
update furniture_items
set definition_id=27,updated_at=now()
where id=429106 and room_id=111;

insert into furniture_items(id,definition_id,owner_player_id,room_id,x,y,z,rotation,extra_data)
overriding system value
values
    (421900,3683,1,111,9,5,0,0,'0'),
    (421901,3681,1,111,9,5,0.65,0,'0')
on conflict(id) do update set
    definition_id=excluded.definition_id,
    owner_player_id=excluded.owner_player_id,
    room_id=excluded.room_id,
    x=excluded.x,
    y=excluded.y,
    z=excluded.z,
    rotation=excluded.rotation,
    extra_data=excluded.extra_data,
    updated_at=now();

insert into room_wired_settings(item_id,int_params,string_param,selection_mode,delay_pulses)
values
    (421900,'[]','',0,0),
    (421901,'[]','Conditions lab: the snapshot target is the RIGHTMOST Color Wheel (ID 429106), not either floor switch. Toggle, move or rotate that wheel before using snapshotchanged.',0,0)
on conflict(item_id) do update set
    int_params=excluded.int_params,
    string_param=excluded.string_param,
    selection_mode=excluded.selection_mode,
    delay_pulses=excluded.delay_pulses,
    updated_at=now();
