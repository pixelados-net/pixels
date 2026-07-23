--liquibase formatted sql

--changeset pixels:pixels-pet-seed-development-0001-pet-qa context:development
insert into pet_product_rules(furniture_definition_id,kind,type_id,energy_delta,happiness_delta,experience_delta,consumable,enabled) values
    (1531,'nest',-1,0,10,0,false,true),
    (1532,'food',-1,25,2,1,true,true),
    (1535,'drink',-1,20,2,1,true,true),
    (3912,'toy',-1,0,20,2,false,true),
    (4221,'saddle',15,0,0,0,true,true),
    (4578,'revive',16,0,0,0,true,true),
    (4582,'seed',16,0,0,0,true,true),
    (4592,'rebreed',16,0,0,0,true,true),
    (4593,'speed',16,0,0,0,true,true),
    (4607,'rebreed',16,0,0,0,true,true),
    (4608,'rebreed',16,0,0,0,true,true),
    (4823,'toy',-1,0,15,1,false,true),
    (4825,'nest',-1,0,0,0,false,true),
    (7977,'package',0,0,0,0,true,true)
on conflict(furniture_definition_id) do update set kind=excluded.kind,type_id=excluded.type_id,
    energy_delta=excluded.energy_delta,happiness_delta=excluded.happiness_delta,
    experience_delta=excluded.experience_delta,consumable=excluded.consumable,enabled=true;

insert into pets(id,owner_player_id,name,type_id,breed_id,palette_id,color,rarity,level,experience,energy,happiness,respect,stats_at,room_id,x,y,z,rotation,posture,has_saddle,can_breed,public_ride,public_breed,grow_at,die_at,state,created_at,updated_at,version)
overriding system value values
    (500001,1,'Pixel',0,1,1,'D5B35B',0,1,0,100,100,0,now(),null,null,null,null,null,'std',false,true,false,false,null,null,'inventory',now()-interval '4 days',now(),1),
    (500002,1,'Rex',0,2,2,'7A4E2D',1,12,540,75,80,5,now(),120,8,8,0,2,'std',false,true,false,true,null,null,'room',now()-interval '20 days',now(),1),
    (500003,3,'Bobby',0,1,1,'D5B35B',0,4,70,90,90,0,now(),120,10,8,0,6,'std',false,true,false,false,null,null,'room',now()-interval '6 days',now(),1),
    (500004,1,'Needs',0,1,1,'D5B35B',0,5,100,25,30,0,now(),121,8,8,0,2,'std',false,true,false,false,null,null,'room',now()-interval '8 days',now(),1),
    (500005,1,'Spirit',15,1,1,'D5B35B',0,10,400,100,100,2,now(),122,8,8,0,2,'std',false,true,false,false,null,null,'room',now()-interval '12 days',now(),1),
    (500006,1,'ParentA',0,1,1,'D5B35B',0,8,250,100,100,0,now(),123,7,8,0,2,'std',false,true,false,true,null,null,'room',now()-interval '10 days',now(),1),
    (500007,2,'ParentB',0,2,2,'7A4E2D',1,8,250,100,100,0,now(),123,9,8,0,6,'std',false,true,false,true,null,null,'room',now()-interval '10 days',now(),1),
    (500008,1,'Sprout',16,1,1,'68A83B',0,1,0,100,100,0,now(),124,6,8,0,2,'std',false,true,false,false,now()+interval '7 days',now()+interval '14 days','room',now()-interval '1 day',now(),1),
    (500009,1,'Bloom',16,1,1,'68A83B',0,3,30,100,100,0,now(),124,8,8,0,2,'std',false,true,false,true,now()-interval '1 day',now()+interval '6 days','room',now()-interval '10 days',now(),1),
    (500010,1,'Wilted',16,1,1,'68A83B',0,3,30,100,100,0,now(),124,10,8,0,2,'std',false,false,false,false,now()-interval '8 days',now()-interval '1 day','room',now()-interval '16 days',now(),1)
on conflict(id) do update set owner_player_id=excluded.owner_player_id,name=excluded.name,type_id=excluded.type_id,
    breed_id=excluded.breed_id,palette_id=excluded.palette_id,color=excluded.color,rarity=excluded.rarity,
    level=excluded.level,experience=excluded.experience,energy=excluded.energy,happiness=excluded.happiness,
    respect=excluded.respect,stats_at=excluded.stats_at,room_id=excluded.room_id,x=excluded.x,y=excluded.y,z=excluded.z,
    rotation=excluded.rotation,posture=excluded.posture,has_saddle=excluded.has_saddle,can_breed=excluded.can_breed,
    public_ride=excluded.public_ride,public_breed=excluded.public_breed,grow_at=excluded.grow_at,die_at=excluded.die_at,
    state=excluded.state,created_at=excluded.created_at,updated_at=now(),deleted_at=null,version=excluded.version;

insert into furniture_items(id,definition_id,owner_player_id,room_id,x,y,z,rotation,extra_data)
overriding system value values
    (500100,1532,1,121,5,3,0,0,'0'),(500101,1535,1,121,6,3,0,0,'0'),
    (500102,4823,1,121,7,3,0,0,'0'),(500103,3912,1,121,8,3,0,0,'0'),
    (500104,1531,1,121,9,3,0,0,'0'),(500110,4221,1,122,5,3,0,0,'0'),
    (500120,4825,1,123,7,3,0,0,'0'),(500121,7977,1,123,5,3,0,0,'0'),
    (500130,4593,1,124,5,3,0,0,'0'),(500131,4578,1,124,6,3,0,0,'0'),
    (500132,4592,1,124,7,3,0,0,'0'),(500133,4582,1,124,9,3,0,0,'0')
on conflict(id) do update set definition_id=excluded.definition_id,owner_player_id=excluded.owner_player_id,
    room_id=excluded.room_id,x=excluded.x,y=excluded.y,z=excluded.z,rotation=excluded.rotation,
    extra_data=excluded.extra_data,deleted_at=null,updated_at=now(),version=1;

select setval(pg_get_serial_sequence('pets','id'),greatest((select max(id) from pets),1));
select setval(pg_get_serial_sequence('furniture_items','id'),greatest((select max(id) from furniture_items),1));
--rollback delete from furniture_items where id between 500100 and 500133; delete from pets where id between 500001 and 500010; delete from pet_product_rules where furniture_definition_id in (1531,1532,1535,3912,4221,4578,4582,4592,4593,4607,4608,4823,4825,7977);
