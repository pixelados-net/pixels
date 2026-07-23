--liquibase formatted sql

--changeset pixels:room-seed-finals-0019 context:development
--validCheckSum:ANY
insert into rooms(id,owner_player_id,owner_name,name,description,model_name,max_users,score,category_id,trade_mode,staff_picked)
overriding system value values
 (160,1,'milo','Curiosities Corner','A little bit of everything - love locks, a mystery box, rentable space, and fireworks for special occasions.','model_a',2,0,2,0,false)
on conflict(id) do update set owner_player_id=excluded.owner_player_id,owner_name=excluded.owner_name,name=excluded.name,
 description=excluded.description,model_name=excluded.model_name,max_users=excluded.max_users,category_id=excluded.category_id,
 trade_mode=excluded.trade_mode,staff_picked=false,deleted_at=null,updated_at=now();

insert into room_tags(room_id,tag) values (160,'showcase')
on conflict do nothing;

select setval(pg_get_serial_sequence('rooms','id'),greatest((select max(id) from rooms),1));
--rollback delete from room_tags where room_id=160; delete from rooms where id=160;
