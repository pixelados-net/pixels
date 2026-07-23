--liquibase formatted sql

--changeset pixels:room-seed-football-qa-entrance-0018 context:development
update rooms set
 description='Inicia 02:00. Ponte detrás del balón y empújalo recto al arco rival. El gol suma y devuelve el balón al centro.',
 updated_at=now(),version=version+1
where id=152;

update room_custom_layouts set door_x=10,door_y=12,door_direction=0,updated_at=now()
where room_id=152;

--rollback update rooms set description='Cancha horizontal: inicia el contador y empuja la pelota por el frente del arco.',updated_at=now(),version=version+1 where id=152; update room_custom_layouts set door_x=1,door_y=7,door_direction=2,updated_at=now() where room_id=152;
