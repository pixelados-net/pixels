--liquibase formatted sql

--changeset pixels:pixels-furniture-seed-development-0032-pet-products context:development
-- Definitions preserve the Arcturus BaseDB identifiers and Nitro sprite identifiers.
insert into furniture_definitions (id,sprite_id,name,public_name,kind,width,length,stack_height,allow_stack,allow_walk,allow_sit,allow_lay,allow_inventory_stack,allow_trade,allow_marketplace_sale,interaction_type,interaction_modes_count,multiheight,custom_params,metadata)
overriding system value values
    (1531,1531,'nest','Pet Nest','floor',1,1,0.10,true,true,false,false,true,true,false,'nest',1,'','','{}'),
    (1532,1532,'petfood1','Doggy Bones','floor',1,1,0.10,true,true,false,false,true,true,false,'pet_food',4,'','','{}'),
    (1535,1535,'waterbowl*4','Blue Water Bowl','floor',1,1,0.10,true,true,false,false,true,true,false,'pet_drink',6,'','','{}'),
    (3912,3912,'pet_toy_trampoline','Pet Trampoline','floor',1,1,1.00,true,true,false,false,true,true,false,'pet_toy',2,'','','{}'),
    (4221,4221,'horse_saddle1','Horse Saddle','floor',1,1,1.00,true,false,false,false,true,true,false,'default',1,'','15 4 9 0 77','{}'),
    (4578,4578,'mnstr_revival','Plant Revival Potion','floor',1,1,1.00,true,false,false,false,true,true,false,'default',2,'','16','{}'),
    (4582,4582,'mnstr_seed','Plant Seed','floor',1,1,1.00,true,false,false,false,true,true,false,'monsterplant_seed',2,'','','{}'),
    (4592,4592,'mnstr_rebreed','Rebreeding Potion 1','floor',1,1,1.00,true,false,false,false,true,true,false,'default',2,'','','{}'),
    (4593,4593,'mnstr_fert','Monster Plant Fertiliser','floor',1,1,1.00,true,false,false,false,true,true,false,'default',2,'','16','{}'),
    (4607,4607,'mnstr_rebreed_2','Rebreeding Potion 2','floor',1,1,1.00,true,false,false,false,true,true,false,'default',2,'','','{}'),
    (4608,4608,'mnstr_rebreed_3','Rebreeding Potion 3','floor',1,1,1.00,true,false,false,false,true,true,false,'default',2,'','','{}'),
    (4823,4823,'pet_toy_ball','Pet Ball','floor',1,1,0.00,true,true,false,false,true,true,false,'pet_toy',1,'','','{}'),
    (4825,4825,'pet_breeding_terrier','Pet Breeding Nest','floor',1,1,1.00,true,false,false,false,true,true,false,'breeding_nest',2,'','','{}'),
    (7977,7977,'petbox_epic','Epic Pet Package','floor',1,1,0.00,true,false,false,false,true,true,false,'pet_package',2,'','0','{}')
on conflict(id) do update set sprite_id=excluded.sprite_id,name=excluded.name,public_name=excluded.public_name,
    kind=excluded.kind,width=excluded.width,length=excluded.length,stack_height=excluded.stack_height,
    allow_stack=excluded.allow_stack,allow_walk=excluded.allow_walk,allow_inventory_stack=excluded.allow_inventory_stack,
    allow_trade=excluded.allow_trade,allow_marketplace_sale=excluded.allow_marketplace_sale,
    interaction_type=excluded.interaction_type,interaction_modes_count=excluded.interaction_modes_count,
    custom_params=excluded.custom_params,metadata=excluded.metadata,deleted_at=null,updated_at=now();

select setval(pg_get_serial_sequence('furniture_definitions','id'),greatest((select max(id) from furniture_definitions),1));
--rollback delete from furniture_definitions where id in (1531,1532,1535,3912,4221,4578,4582,4592,4593,4607,4608,4823,4825,7977);
