--liquibase formatted sql

--changeset pixels:pixels-catalog-seed-development-0023-pets context:development
insert into catalog_pages (id,parent_id,name,layout,icon_color,icon_image,order_num,visible,enabled,club_only,new_additions)
overriding system value values
    (109,1,'pets','pets3',2,5,9,true,true,false,true),
    (110,109,'pet_dogs','pets',2,5,1,true,true,false,false),
    (111,109,'pet_cats','pets',2,5,2,true,true,false,false),
    (112,109,'pet_horses','pets',2,5,3,true,true,false,false),
    (113,109,'pet_accessories','default_3x3',2,5,4,true,true,false,false),
    (114,109,'pet_plants_breeding','default_3x3',2,5,5,true,true,false,false)
on conflict(id) do update set parent_id=excluded.parent_id,name=excluded.name,layout=excluded.layout,
    icon_color=excluded.icon_color,icon_image=excluded.icon_image,order_num=excluded.order_num,
    visible=true,enabled=true,club_only=false,new_additions=excluded.new_additions;

insert into catalog_items (id,page_id,definition_id,name,cost_credits,cost_points,points_type,amount,limited_stack,limited_sells,bundle_discount_enabled,giftable,club_only,order_num,enabled,extra_data,reward_kind,pet_type_id,pet_product_code)
overriding system value values
    (1301,110,null,'pet_dog',20,0,-1,1,0,0,false,false,false,1,true,'','pet',0,'pet0'),
    (1302,111,null,'pet_cat',20,0,-1,1,0,0,false,false,false,1,true,'','pet',1,'pet1'),
    (1303,112,null,'pet_horse',35,0,-1,1,0,0,false,false,false,1,true,'','pet',15,'pet15'),
    (1310,113,1532,'pet_food_bones',3,0,-1,1,0,0,false,false,false,1,true,'0','furniture',null,''),
    (1311,113,1535,'pet_water_bowl',3,0,-1,1,0,0,false,false,false,2,true,'0','furniture',null,''),
    (1312,113,4823,'pet_toy_ball',5,0,-1,1,0,0,false,false,false,3,true,'0','furniture',null,''),
    (1313,113,3912,'pet_toy_trampoline',8,0,-1,1,0,0,false,false,false,4,true,'0','furniture',null,''),
    (1314,113,1531,'pet_nest',6,0,-1,1,0,0,false,false,false,5,true,'0','furniture',null,''),
    (1315,113,4221,'pet_horse_saddle',12,0,-1,1,0,0,false,false,false,6,true,'0','furniture',null,''),
    (1320,114,4825,'pet_breeding_nest',15,0,-1,1,0,0,false,false,false,1,true,'0','furniture',null,''),
    (1321,114,7977,'pet_package_dog',18,0,-1,1,0,0,false,false,false,2,true,'0','furniture',null,''),
    (1322,114,4582,'monsterplant_seed',8,0,-1,1,0,0,false,false,false,3,true,'0','furniture',null,''),
    (1323,114,4593,'monsterplant_speed',6,0,-1,1,0,0,false,false,false,4,true,'0','furniture',null,''),
    (1324,114,4578,'monsterplant_revive',8,0,-1,1,0,0,false,false,false,5,true,'0','furniture',null,''),
    (1325,114,4592,'monsterplant_rebreed',8,0,-1,1,0,0,false,false,false,6,true,'0','furniture',null,'')
on conflict(id) do update set page_id=excluded.page_id,definition_id=excluded.definition_id,name=excluded.name,
    cost_credits=excluded.cost_credits,cost_points=excluded.cost_points,points_type=excluded.points_type,
    amount=excluded.amount,limited_stack=0,limited_sells=0,bundle_discount_enabled=false,giftable=false,
    club_only=false,order_num=excluded.order_num,enabled=true,extra_data=excluded.extra_data,
    reward_kind=excluded.reward_kind,pet_type_id=excluded.pet_type_id,pet_product_code=excluded.pet_product_code,
    deleted_at=null,updated_at=now();

select setval(pg_get_serial_sequence('catalog_pages','id'),greatest((select max(id) from catalog_pages),1));
select setval(pg_get_serial_sequence('catalog_items','id'),greatest((select max(id) from catalog_items),1));
--rollback delete from catalog_items where id between 1301 and 1325; delete from catalog_pages where id between 109 and 114;
