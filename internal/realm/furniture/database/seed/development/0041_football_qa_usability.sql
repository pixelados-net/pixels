--liquibase formatted sql

--changeset pixels:furniture-seed-football-qa-usability-0041 context:development
update furniture_definitions set allow_walk=false,updated_at=now()
where id in (950019,950020);

update furniture_definitions set custom_params='120,180,300,600',updated_at=now()
where id=950024;

update furniture_items set
 x=case id when 961400 then 3 when 961401 then 17 when 961402 then 10 else x end,
 y=case id when 961400 then 6 when 961401 then 6 when 961402 then 7 else y end,
 rotation=case id when 961400 then 6 when 961401 then 2 when 961402 then 0 else rotation end,
 extra_data=case when id in (961403,961404) then '0' when id=961406 then '120' else extra_data end,
 updated_at=now(),version=version+1
where room_id=152 and id between 961400 and 961406;

--rollback update furniture_definitions set allow_walk=true,updated_at=now() where id in (950019,950020); update furniture_definitions set custom_params='30,60,120,180,300,600',updated_at=now() where id=950024; update furniture_items set y=7,rotation=case id when 961400 then 6 when 961401 then 2 else rotation end,extra_data=case when id=961406 then '30' else extra_data end,updated_at=now(),version=version+1 where room_id=152 and id between 961400 and 961406;
