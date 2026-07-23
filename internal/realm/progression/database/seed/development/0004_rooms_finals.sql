--liquibase formatted sql

--changeset pixels:progression-seed-rooms-finals-0004 context:development
insert into achievement_definitions(id,name,category,subcategory,trigger_key,visible,enabled)
overriding system value values
 (930064,'FireworksCharger','room_builder','furniture','furniture.firework.charged',true,true)
on conflict(id) do update set name=excluded.name,category=excluded.category,subcategory=excluded.subcategory,
 trigger_key=excluded.trigger_key,visible=excluded.visible,enabled=excluded.enabled,updated_at=now();

insert into achievement_levels(definition_id,level,progress_needed,reward_currency_type,reward_amount,score_points)
select 930064,level,threshold,0,level*5,greatest(5,(level-1)*5)
from unnest(array[1,5,20,50,100]::bigint[]) with ordinality valueset(threshold,level)
on conflict(definition_id,level) do update set progress_needed=excluded.progress_needed,reward_amount=excluded.reward_amount,score_points=excluded.score_points;

select setval(pg_get_serial_sequence('achievement_definitions','id'),greatest((select max(id) from achievement_definitions),1));
--rollback delete from achievement_definitions where id=930064;
