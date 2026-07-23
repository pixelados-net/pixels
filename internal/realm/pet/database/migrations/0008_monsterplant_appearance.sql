--liquibase formatted sql

--changeset pixels:pixels-pet-0008-monsterplant-appearance
insert into pet_appearance_parts(pet_id,ordinal,layer_id,part_id,palette_id)
select p.id,g.layer,g.layer,
    case when g.layer=0 then -1 else 1+mod(p.id*31+g.layer*17,12)::integer end,
    case when g.layer=0 then 10 else mod(p.id*13+g.layer*7,11)::integer end
from pets p cross join generate_series(0,4) as g(layer)
where p.type_id=16 and p.deleted_at is null
  and not exists(select 1 from pet_appearance_parts existing where existing.pet_id=p.id)
on conflict(pet_id,ordinal) do nothing;
--rollback delete from pet_appearance_parts a where a.pet_id in (select p.id from pets p join pet_appearance_parts candidate on candidate.pet_id=p.id where p.type_id=16 group by p.id having count(*)=5 and bool_and(candidate.ordinal between 0 and 4 and candidate.layer_id=candidate.ordinal and candidate.part_id=case when candidate.ordinal=0 then -1 else 1+mod(p.id*31+candidate.ordinal*17,12)::integer end and candidate.palette_id=case when candidate.ordinal=0 then 10 else mod(p.id*13+candidate.ordinal*7,11)::integer end));
