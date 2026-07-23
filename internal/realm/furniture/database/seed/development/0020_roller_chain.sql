--liquibase formatted sql

--changeset pixels:pixels-furniture-seed-development-0020-roller-chain context:development
-- A three-tile eastbound chain in the seeded lobby with one table mounted on its first belt.
insert into furniture_items (definition_id,owner_player_id,room_id,x,y,z,rotation,extra_data)
select values_row.definition_id,1,1,values_row.x,12,values_row.z,2,'0'
from (values (7,5,0.00::numeric),(7,6,0.00::numeric),(7,7,0.00::numeric),(3,5,0.50::numeric))
as values_row(definition_id,x,z)
where not exists (
    select 1 from furniture_items existing
    where existing.room_id=1 and existing.definition_id=values_row.definition_id
      and existing.x=values_row.x and existing.y=12 and existing.z=values_row.z
      and existing.deleted_at is null
);
--rollback delete from furniture_items where room_id=1 and y=12 and ((definition_id=7 and x between 5 and 7 and z=0.00) or (definition_id=3 and x=5 and z=0.50));
