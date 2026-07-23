--liquibase formatted sql

--changeset pixels:room-seed-games-0016 context:development
--validCheckSum:ANY
insert into rooms(id,owner_player_id,owner_name,name,description,model_name,max_users,score,category_id,trade_mode,staff_picked)
overriding system value values
 (150,1,'milo','Battle Banzai Arena','Battle Banzai teams, tiles, puck, sphere and timer.','model_a',25,0,2,0,false),
 (151,1,'milo','Freeze Arena','Freeze teams, blocks, exits, power-ups and timer.','model_a',25,0,2,0,false),
 (152,1,'milo','Football Pitch','Football physics, directional goals and scoreboards.','model_a',25,0,2,0,false),
 (153,1,'milo','Tag & Rollerskate Arena','IceTag, Bunnyrun, rollerskate tag, and a quick poll on the way in.','model_a',25,0,2,0,false)
on conflict(id) do update set owner_player_id=excluded.owner_player_id,owner_name=excluded.owner_name,name=excluded.name,
 description=excluded.description,model_name=excluded.model_name,max_users=excluded.max_users,category_id=excluded.category_id,
 trade_mode=excluded.trade_mode,staff_picked=false,deleted_at=null,updated_at=now();

insert into room_tags(room_id,tag) values
 (150,'banzai'),(151,'freeze'),(152,'football'),(153,'tag')
on conflict do nothing;

select setval(pg_get_serial_sequence('rooms','id'),greatest((select max(id) from rooms),1));
--rollback delete from room_tags where room_id between 150 and 153; delete from rooms where id between 150 and 153;
