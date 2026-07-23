--liquibase formatted sql

--changeset pixels:pixels-pet-seed-development-0002-monsterplant-appearance context:development
delete from pet_appearance_parts where pet_id in (500008,500009,500010);
insert into pet_appearance_parts(pet_id,ordinal,layer_id,part_id,palette_id)
select genetics.pet_id,genetics.ordinal,genetics.layer_id,genetics.part_id,genetics.palette_id
from (values
    (500008,0,0,-1,10),(500008,1,1,1,3),(500008,2,2,2,5),(500008,3,3,3,2),(500008,4,4,4,6),
    (500009,0,0,-1,10),(500009,1,1,6,4),(500009,2,2,8,1),(500009,3,3,10,7),(500009,4,4,12,9),
    (500010,0,0,-1,10),(500010,1,1,12,2),(500010,2,2,1,10),(500010,3,3,7,4),(500010,4,4,9,8)
) as genetics(pet_id,ordinal,layer_id,part_id,palette_id)
join pets p on p.id=genetics.pet_id
on conflict(pet_id,ordinal) do update set layer_id=excluded.layer_id,part_id=excluded.part_id,palette_id=excluded.palette_id;
--rollback delete from pet_appearance_parts where pet_id between 500008 and 500010;
