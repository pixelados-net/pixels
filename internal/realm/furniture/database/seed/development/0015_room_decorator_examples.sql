--liquibase formatted sql

--changeset pixels:pixels-furniture-seed-development-0015-room-decorator-examples context:development
--validCheckSum: 9:a16d8f34eae37b6fbbfa6d4c9d163aad
delete from furniture_items where id between 21100 and 21104;
insert into furniture_items(id,definition_id,owner_player_id,room_id,x,y,z,rotation,wall_position,extra_data,metadata)
overriding system value values
    (21100,9,1,100,null,null,null,0,':w=3,2 l=1,1 r','2,1,1,#74F5F5,180','{"seed":"decorator"}'),
    (21101,20,1,100,null,null,null,0,':w=4,2 l=1,1 r','FFFF33 Welcome to Pixels!','{"seed":"decorator"}'),
    (21102,40,1,100,5,4,0,0,null,'0','{"seed":"decorator"}'),
    (21103,41,1,100,6,4,0,0,null,'{"gender":"M","figure":"hd-99999-99998.ch-210-66.lg-270-82.sh-290-80","name":"Starter look"}','{"seed":"decorator"}'),
    (21104,42,1,100,7,4,0,0,null,'0:0:0:0','{"seed":"decorator"}');
select setval(pg_get_serial_sequence('furniture_items','id'),greatest((select max(id) from furniture_items),1));
--rollback delete from furniture_items where id between 21100 and 21104;
