--liquibase formatted sql

--changeset pixels:room-seed-games-layouts-0017 context:development
update rooms set
 name=case id when 150 then 'QA · Battle Banzai' when 151 then 'QA · Freeze' when 152 then 'QA · Football' else 'QA · Tag y Polls' end,
 description=case id
  when 150 then 'Tablero 8x8: elige un gate, inicia el contador y pisa tres veces cada tile.'
  when 151 then 'Laberinto 9x7: elige equipo, inicia el contador y usa doble clic en tile o bloque.'
  when 152 then 'Cancha horizontal: inicia el contador y empuja la pelota por el frente del arco.'
  else 'Tres zonas separadas: IceTag, Rollerskate y Bunnyrun; el poll se ofrece al entrar.' end,
 model_name='model_a',max_users=16,updated_at=now(),version=version+1
where id between 150 and 153;

insert into room_custom_layouts(room_id,heightmap,door_x,door_y,door_direction,wall_thickness,floor_thickness,wall_height) values
 (150,E'xxxxxxxxxxxxxxxxxxxx\r\nx000000000000000000x\r\nx000000000000000000x\r\nx000000000000000000x\r\nx000000000000000000x\r\nx000000000000000000x\r\nx000000000000000000x\r\nx000000000000000000x\r\nx000000000000000000x\r\nx000000000000000000x\r\nx000000000000000000x\r\nx000000000000000000x\r\nx000000000000000000x\r\nx000000000000000000x\r\nx000000000000000000x\r\nxxxxxxxxxxxxxxxxxxxx',1,8,2,1,0,-1),
 (151,E'xxxxxxxxxxxxxxxxxxxx\r\nx000000000000000000x\r\nx000000000000000000x\r\nx000000000000000000x\r\nx000000000000000000x\r\nx000000000000000000x\r\nx000000000000000000x\r\nx000000000000000000x\r\nx000000000000000000x\r\nx000000000000000000x\r\nx000000000000000000x\r\nx000000000000000000x\r\nx000000000000000000x\r\nx000000000000000000x\r\nx000000000000000000x\r\nxxxxxxxxxxxxxxxxxxxx',1,8,2,1,0,-1),
 (152,E'xxxxxxxxxxxxxxxxxxxxxx\r\nx00000000000000000000x\r\nx00000000000000000000x\r\nx00000000000000000000x\r\nx00000000000000000000x\r\nx00000000000000000000x\r\nx00000000000000000000x\r\nx00000000000000000000x\r\nx00000000000000000000x\r\nx00000000000000000000x\r\nx00000000000000000000x\r\nx00000000000000000000x\r\nx00000000000000000000x\r\nxxxxxxxxxxxxxxxxxxxxxx',1,7,2,1,0,-1),
 (153,E'xxxxxxxxxxxxxxxxxxxxxxxx\r\nx0000000000000000000000x\r\nx0000000000000000000000x\r\nx0000000000000000000000x\r\nx0000000000000000000000x\r\nx0000000000000000000000x\r\nx0000000000000000000000x\r\nx0000000000000000000000x\r\nx0000000000000000000000x\r\nx0000000000000000000000x\r\nx0000000000000000000000x\r\nx0000000000000000000000x\r\nx0000000000000000000000x\r\nx0000000000000000000000x\r\nx0000000000000000000000x\r\nxxxxxxxxxxxxxxxxxxxxxxxx',1,8,2,1,0,-1)
on conflict(room_id) do update set heightmap=excluded.heightmap,door_x=excluded.door_x,door_y=excluded.door_y,
 door_direction=excluded.door_direction,wall_thickness=excluded.wall_thickness,
 floor_thickness=excluded.floor_thickness,wall_height=excluded.wall_height,updated_at=now();

--rollback delete from room_custom_layouts where room_id between 150 and 153;
