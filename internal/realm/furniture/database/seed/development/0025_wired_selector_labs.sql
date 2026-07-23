--liquibase formatted sql

--changeset pixels:pixels-furniture-seed-development-0025-wired-selector-labs context:development
-- Random and unseen selectors own independent stacks so their behavior is observable and deterministic.
update furniture_items
set x = 6, y = 8, z = 1.30
where id = 411004;

insert into furniture_items(id,definition_id,owner_player_id,room_id,x,y,z,rotation,extra_data)
overriding system value
values
    (411006,3671,1,111,6,8,0,0,'0'),
    (411007,3681,1,111,6,8,0.65,0,'0'),
    (411008,3681,1,111,6,8,1.95,0,'0')
on conflict(id) do update set definition_id=excluded.definition_id,owner_player_id=excluded.owner_player_id,
    room_id=excluded.room_id,x=excluded.x,y=excluded.y,z=excluded.z,rotation=excluded.rotation,extra_data=excluded.extra_data;

insert into room_wired_settings(item_id,int_params,string_param,selection_mode,delay_pulses)
values
    (411006,'[2]','',0,0),
    (411007,'[]','Unseen branch A.',0,0),
    (411008,'[]','Unseen branch B.',0,0)
on conflict(item_id) do update set int_params=excluded.int_params,string_param=excluded.string_param,
    selection_mode=excluded.selection_mode,delay_pulses=excluded.delay_pulses,updated_at=now();

select setval(pg_get_serial_sequence('furniture_items','id'),greatest((select max(id) from furniture_items),1));
--rollback delete from furniture_items where id between 411006 and 411008;
--rollback update furniture_items set x=4,y=2,z=2.60 where id=411004;
