--liquibase formatted sql

--changeset pixels:furniture-seed-football-goal-orientation-0042 context:development
update furniture_items set
 rotation=case id when 961400 then 2 when 961401 then 6 else rotation end,
 updated_at=now(),version=version+1
where room_id=152 and id in (961400,961401);

--rollback update furniture_items set rotation=case id when 961400 then 6 when 961401 then 2 else rotation end,updated_at=now(),version=version+1 where room_id=152 and id in (961400,961401);
