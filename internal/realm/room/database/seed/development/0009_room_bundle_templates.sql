--liquibase formatted sql

--changeset pixels:pixels-room-seed-development-0009-room-bundle-templates context:development
--validCheckSum:ANY
insert into rooms (id,owner_player_id,owner_name,name,description,model_name,max_users,category_id,trade_mode,is_bundle_template)
overriding system value values
    (100,1,'milo','Starter Loft','A complete social loft room bundle.','model_a',25,3,2,true),
    (101,1,'milo','Interactive Lounge','A custom-layout games room bundle.','model_b',25,2,2,true),
    (102,1,'milo','Cozy Bedroom','A furnished bedroom room bundle.','model_c',15,3,0,true)
on conflict (id) do update set
    name=excluded.name,description=excluded.description,model_name=excluded.model_name,
    max_users=excluded.max_users,category_id=excluded.category_id,trade_mode=excluded.trade_mode,
    is_bundle_template=true,deleted_at=null,updated_at=now(),version=rooms.version+1;

insert into room_custom_layouts (room_id,heightmap,door_x,door_y,door_direction,wall_thickness,floor_thickness,wall_height)
values (101,E'xxxxxxxxxxxx\rxxxxx0000000\rxxxxx0000000\rxxxxx0000000\rxxxxx1110000\r000001110000\rx00001110000\rx00000000000\rx00000000000\rx00000000000\rx00000000000\rxxxxxxxxxxxx\rxxxxxxxxxxxx\rxxxxxxxxxxxx\rxxxxxxxxxxxx\rxxxxxxxxxxxx',0,5,2,1,1,-1)
on conflict (room_id) do update set heightmap=excluded.heightmap,door_x=excluded.door_x,door_y=excluded.door_y,
    door_direction=excluded.door_direction,wall_thickness=excluded.wall_thickness,
    floor_thickness=excluded.floor_thickness,wall_height=excluded.wall_height,updated_at=now();

select setval(pg_get_serial_sequence('rooms','id'),greatest((select max(id) from rooms),1));
--rollback delete from room_custom_layouts where room_id=101; delete from rooms where id in (100,101,102);
