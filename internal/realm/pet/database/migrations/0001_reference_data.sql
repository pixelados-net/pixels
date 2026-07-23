--liquibase formatted sql
--changeset pixels:pixels-pet-0001-reference-data
create table pet_species (
    type_id integer primary key check (type_id between 0 and 35),
    slug text not null unique,
    display_key text not null,
    behavior_kind text not null default 'generic',
    max_level integer not null default 20 check (max_level between 1 and 20),
    rideable boolean not null default false,
    breedable boolean not null default true,
    plant boolean not null default false,
    enabled boolean not null default true,
    version bigint not null default 1
);

create table pet_breeds (
    type_id integer not null references pet_species(type_id),
    breed_id integer not null check (breed_id >= 0),
    palette_id integer not null check (palette_id >= 0),
    color text not null check (color ~ '^[0-9A-F]{6}$'),
    sellable boolean not null default true,
    rarity integer not null default 0 check (rarity >= 0),
    enabled boolean not null default true,
    primary key (type_id,breed_id,palette_id)
);

create table pet_commands (
    id integer primary key check (id between 0 and 46 and id <> 39),
    name_key text not null unique,
    required_level integer not null check (required_level between 1 and 20),
    family text not null,
    energy_cost integer not null default 0 check (energy_cost >= 0),
    experience_reward integer not null default 0 check (experience_reward >= 0),
    duration_ms bigint not null default 1500 check (duration_ms between 0 and 60000),
    cooldown_ms bigint not null default 1000 check (cooldown_ms between 0 and 600000),
    enabled boolean not null default true
);

create table pet_species_commands (
    type_id integer not null references pet_species(type_id),
    command_id integer not null references pet_commands(id),
    primary key (type_id,command_id)
);

create table pet_product_rules (
    furniture_definition_id bigint primary key references furniture_definitions(id),
    kind text not null check (kind in ('food','drink','toy','nest','saddle','revive','rebreed','speed','seed','package')),
    type_id integer not null default -1 check (type_id between -1 and 35),
    energy_delta integer not null default 0,
    happiness_delta integer not null default 0,
    experience_delta integer not null default 0,
    consumable boolean not null default true,
    enabled boolean not null default true
);

insert into pet_species(type_id,slug,display_key,behavior_kind,rideable,breedable,plant,enabled) values
(0,'dog','pet.species.dog','generic',false,true,false,true),(1,'cat','pet.species.cat','generic',false,true,false,true),
(2,'crocodile','pet.species.crocodile','generic',false,true,false,true),(3,'terrier','pet.species.terrier','generic',false,true,false,true),
(4,'bear','pet.species.bear','generic',false,true,false,true),(5,'pig','pet.species.pig','generic',false,true,false,true),
(6,'lion','pet.species.lion','generic',false,true,false,true),(7,'rhino','pet.species.rhino','generic',false,true,false,true),
(8,'spider','pet.species.spider','generic',false,true,false,true),(9,'turtle','pet.species.turtle','generic',false,true,false,true),
(10,'chicken','pet.species.chicken','generic',false,true,false,true),(11,'frog','pet.species.frog','generic',false,true,false,true),
(12,'dragon','pet.species.dragon','generic',false,true,false,true),(13,'reserved13','pet.species.reserved13','disabled',false,false,false,false),
(14,'monkey','pet.species.monkey','generic',false,true,false,true),(15,'horse','pet.species.horse','horse',true,true,false,true),
(16,'monsterplant','pet.species.monsterplant','plant',false,true,true,true),(17,'bunny','pet.species.bunny','generic',false,true,false,true),
(18,'bunnyevil','pet.species.bunnyevil','generic',false,true,false,true),(19,'bunnydepressed','pet.species.bunnydepressed','generic',false,true,false,true),
(20,'bunnylove','pet.species.bunnylove','generic',false,true,false,true),(21,'pigeongood','pet.species.pigeongood','generic',false,true,false,true),
(22,'pigeonevil','pet.species.pigeonevil','generic',false,true,false,true),(23,'demonmonkey','pet.species.demonmonkey','generic',false,true,false,true),
(24,'babybear','pet.species.babybear','generic',false,true,false,true),(25,'babyterrier','pet.species.babyterrier','generic',false,true,false,true),
(26,'gnome','pet.species.gnome','generic',false,true,false,true),(27,'leprechaun','pet.species.leprechaun','generic',false,true,false,true),
(28,'kittenbaby','pet.species.kittenbaby','generic',false,true,false,true),(29,'puppybaby','pet.species.puppybaby','generic',false,true,false,true),
(30,'pigletbaby','pet.species.pigletbaby','generic',false,true,false,true),(31,'haloompa','pet.species.haloompa','generic',false,true,false,true),
(32,'fools','pet.species.fools','generic',false,true,false,true),(33,'pterosaur','pet.species.pterosaur','generic',false,true,false,true),
(34,'velociraptor','pet.species.velociraptor','generic',false,true,false,true),(35,'cow','pet.species.cow','generic',false,true,false,true);

insert into pet_breeds(type_id,breed_id,palette_id,color,sellable,rarity,enabled)
select type_id,0,0,'FFFFFF',enabled,0,enabled from pet_species;
insert into pet_breeds(type_id,breed_id,palette_id,color,sellable,rarity,enabled) values
(0,1,1,'D5B35B',true,0,true),(0,2,2,'7A4E2D',true,1,true),(1,1,1,'B8B8B8',true,0,true),
(15,1,1,'D5B35B',true,0,true),(15,2,2,'2E2E2E',true,1,true),(16,1,1,'68A83B',false,0,true);

insert into pet_commands(id,name_key,required_level,family,energy_cost,experience_reward,duration_ms,cooldown_ms)
select id,'pet.command.'||id,least(20,1+(id/3)),case when id in (0,1,2,3,4,5,6,7,8,9) then 'basic' when id in (10,11,12,13,14,15,16,17,18) then 'movement' when id in (19,20,21,22,23,24,25,26,27,28) then 'gesture' else 'advanced' end,1,2,1500,1000
from generate_series(0,46) id where id <> 39;
insert into pet_species_commands(type_id,command_id)
select s.type_id,c.id from pet_species s cross join pet_commands c where s.enabled and c.enabled;
--rollback drop table if exists pet_product_rules; drop table if exists pet_species_commands; drop table if exists pet_commands; drop table if exists pet_breeds; drop table if exists pet_species;
