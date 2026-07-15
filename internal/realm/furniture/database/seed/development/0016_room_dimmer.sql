--liquibase formatted sql

--changeset pixels:pixels-furniture-seed-development-0016-room-dimmer context:development
insert into furniture_definitions (id,sprite_id,name,public_name,kind,width,length,stack_height,allow_stack,allow_walk,allow_sit,allow_lay,allow_inventory_stack,interaction_type,interaction_modes_count,multiheight,custom_params,metadata)
overriding system value values
    (43,202,'teleport_tile','Teleport Tile','floor',1,1,0.00,true,true,false,false,true,'teleport_tile',1,'','','{}')
on conflict(id) do update set sprite_id=excluded.sprite_id,name=excluded.name,public_name=excluded.public_name,kind=excluded.kind,interaction_type=excluded.interaction_type,interaction_modes_count=excluded.interaction_modes_count;
update furniture_items set definition_id=43,updated_at=now() where id in (1005,1006) and definition_id=9;
update furniture_definitions set sprite_id=4027,name='roomdimmer',public_name='Mood Light',kind='wall',width=1,length=1,stack_height=1.00,allow_stack=true,allow_walk=false,allow_sit=false,allow_lay=false,allow_inventory_stack=true,interaction_type='dimmer',interaction_modes_count=1,multiheight='',custom_params='',updated_at=now() where id=9;
select setval(pg_get_serial_sequence('furniture_definitions','id'),greatest((select max(id) from furniture_definitions),1));
--rollback update furniture_items set definition_id=9,updated_at=now() where id in (1005,1006) and definition_id=43; update furniture_definitions set interaction_type='teleport_tile',allow_walk=true,updated_at=now() where id=9; delete from furniture_definitions where id=43;
