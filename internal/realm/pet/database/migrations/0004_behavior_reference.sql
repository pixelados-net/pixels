--liquibase formatted sql
--changeset pixels:pixels-pet-0004-behavior-reference
create table pet_vocals (
    id bigint generated always as identity primary key,
    type_id integer not null references pet_species(type_id),
    mood text not null check (mood in ('idle','happy','hungry','command')),
    text_key text not null,
    weight integer not null default 1 check (weight between 1 and 10000),
    cooldown_ms bigint not null default 15000 check (cooldown_ms between 1000 and 3600000),
    enabled boolean not null default true
);
create index pet_vocals_type_idx on pet_vocals(type_id,id) where enabled;

create table pet_breeding_rules (
    parent_one_type_id integer not null references pet_species(type_id),
    parent_two_type_id integer not null references pet_species(type_id),
    result_type_id integer not null references pet_species(type_id),
    enabled boolean not null default true,
    primary key(parent_one_type_id,parent_two_type_id),
    constraint pet_breeding_parent_order_chk check (parent_one_type_id <= parent_two_type_id)
);

create table pet_breeding_races (
    result_type_id integer not null,
    breed_id integer not null,
    palette_id integer not null,
    weight integer not null default 100 check (weight between 1 and 1000000),
    mutation boolean not null default false,
    enabled boolean not null default true,
    primary key(result_type_id,breed_id,palette_id),
    foreign key(result_type_id,breed_id,palette_id) references pet_breeds(type_id,breed_id,palette_id)
);

insert into pet_vocals(type_id,mood,text_key,weight,cooldown_ms)
select type_id,'idle','pet.vocal.generic',1,15000 from pet_species where enabled;
update pet_vocals set text_key='pet.vocal.dog' where type_id in (0,3,25,29);
update pet_vocals set text_key='pet.vocal.cat' where type_id in (1,28);
update pet_vocals set text_key='pet.vocal.horse' where type_id=15;
update pet_vocals set text_key='pet.vocal.plant' where type_id=16;

insert into pet_breeding_rules(parent_one_type_id,parent_two_type_id,result_type_id)
select type_id,type_id,type_id from pet_species where enabled and breedable;
insert into pet_breeding_races(result_type_id,breed_id,palette_id,weight,mutation,enabled)
select type_id,breed_id,palette_id,greatest(1,100/(rarity+1)),rarity>0,enabled from pet_breeds where enabled;
--rollback drop table if exists pet_breeding_races; drop table if exists pet_breeding_rules; drop table if exists pet_vocals;
