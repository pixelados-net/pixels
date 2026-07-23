--liquibase formatted sql

--changeset pixels:pixels-room-seed-development-0014-pet-qa-rooms context:development
insert into rooms (id,owner_player_id,owner_name,name,description,model_name,max_users,score,category_id,trade_mode,staff_picked,allow_pets,allow_pets_eat)
overriding system value values
    (120,1,'demo','PETS QA Core','Inventory, placement, info, respect, training and movement.','model_a',25,0,2,0,false,true,false),
    (121,1,'demo','PETS QA Needs','Food, drink, toys, nest and AllowPetsEat behavior.','model_a',25,0,2,0,false,true,true),
    (122,1,'demo','PETS QA Riding','Horse saddle, owner/public riding and obstacle corridor.','model_a',25,0,2,0,false,true,false),
    (123,1,'demo','PETS QA Breeding','Packages, compatible parents, breeding nest and cancellation.','model_a',25,0,2,0,false,true,false),
    (124,1,'demo','PETS QA Plants','Growing, mature and dead monsterplants with every supplement.','model_a',25,0,2,0,false,true,false)
on conflict(id) do update set owner_player_id=excluded.owner_player_id,owner_name=excluded.owner_name,
    name=excluded.name,description=excluded.description,model_name=excluded.model_name,max_users=excluded.max_users,
    category_id=excluded.category_id,trade_mode=excluded.trade_mode,staff_picked=false,
    allow_pets=excluded.allow_pets,allow_pets_eat=excluded.allow_pets_eat;

insert into room_tags(room_id,tag) values
    (120,'qa-pets-core'),(121,'qa-pets-needs'),(122,'qa-pets-riding'),
    (123,'qa-pets-breeding'),(124,'qa-pets-plants')
on conflict do nothing;

select setval(pg_get_serial_sequence('rooms','id'),greatest((select max(id) from rooms),1));
--rollback delete from room_tags where room_id between 120 and 124; delete from rooms where id between 120 and 124;
