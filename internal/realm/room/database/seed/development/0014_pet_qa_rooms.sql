--liquibase formatted sql

--changeset pixels:pixels-room-seed-development-0014-pet-qa-rooms context:development
--validCheckSum:ANY
insert into rooms (id,owner_player_id,owner_name,name,description,model_name,max_users,score,category_id,trade_mode,staff_picked,allow_pets,allow_pets_eat)
overriding system value values
    (120,1,'milo','The Pet Corner','Everyday pet care - fetch, play, and a little training.','model_a',25,0,2,0,false,true,false),
    (121,1,'milo','Feeding Grounds','Food, drink, and toys - happy pets live here.','model_a',25,0,2,0,false,true,true),
    (122,1,'milo','Saddle Trail','Saddle up and take your horse for a ride.','model_a',25,0,2,0,false,true,false),
    (123,1,'milo','The Nursery','Where new pets are born and matched with new owners.','model_a',25,0,2,0,false,true,false),
    (124,1,'milo','Monsterplant Greenhouse','Grow, water, and harvest monsterplants.','model_a',25,0,2,0,false,true,false)
on conflict(id) do update set owner_player_id=excluded.owner_player_id,owner_name=excluded.owner_name,
    name=excluded.name,description=excluded.description,model_name=excluded.model_name,max_users=excluded.max_users,
    category_id=excluded.category_id,trade_mode=excluded.trade_mode,staff_picked=false,
    allow_pets=excluded.allow_pets,allow_pets_eat=excluded.allow_pets_eat;

insert into room_tags(room_id,tag) values
    (120,'pets'),(121,'feeding'),(122,'riding'),
    (123,'breeding'),(124,'plants')
on conflict do nothing;

select setval(pg_get_serial_sequence('rooms','id'),greatest((select max(id) from rooms),1));
--rollback delete from room_tags where room_id between 120 and 124; delete from rooms where id between 120 and 124;
