--liquibase formatted sql

--changeset pixels:progression-seed-games-polls-0003 context:development
--validCheckSum:ANY
update achievement_definitions set enabled=true,updated_at=now()
where name in ('BattleBallPlayer','FreezePlayer','FootballPlayer','TagPlayer');

insert into achievement_definitions(id,name,category,subcategory,trigger_key,visible,enabled)
overriding system value values
 (930052,'BattleBallWinner','games','banzai','game.banzai.won',true,true),
 (930053,'BattleBallTilesLocked','games','banzai','game.banzai.tile.locked',true,true),
 (930054,'FreezeWinner','games','freeze','game.freeze.won',true,true),
 (930055,'FreezePowerUp','games','freeze','game.freeze.powerup',true,true),
 (930056,'EsA','games','freeze','game.freeze.player.frozen',true,true),
 (930057,'FootballGoalScored','games','football','game.football.goal',true,true),
 (930058,'TagA','games','tag','game.tag.placed',true,true),
 (930059,'TagC','games','tag','game.tag.minutes',true,true),
 (930060,'RbTagA','games','tag','game.rollerskate.placed',true,true),
 (930061,'RbTagC','games','tag','game.rollerskate.minutes',true,true),
 (930062,'GameAuthorExperience','games','','game.authored',true,true),
 (930063,'GamePlayed','games','','game.played',true,true)
on conflict(id) do update set name=excluded.name,category=excluded.category,subcategory=excluded.subcategory,
 trigger_key=excluded.trigger_key,visible=excluded.visible,enabled=excluded.enabled,updated_at=now();

insert into achievement_levels(definition_id,level,progress_needed,reward_currency_type,reward_amount,score_points)
select definition_id,level,threshold,0,level*5,greatest(5,(level-1)*5)
from (values
 (930052,array[1,5,20,50,100]::bigint[]),(930053,array[1,10,50,200,500]::bigint[]),
 (930054,array[1,5,20,50,100]::bigint[]),(930055,array[1,10,50,200,500]::bigint[]),
 (930056,array[1,10,50,200,500]::bigint[]),(930057,array[1,10,50,200,500]::bigint[]),
 (930058,array[1,5,20,50,100]::bigint[]),(930059,array[10,60,180,600,1200]::bigint[]),
 (930060,array[1,5,20,50,100]::bigint[]),(930061,array[10,60,180,600,1200]::bigint[]),
 (930062,array[1,10,50,200,500]::bigint[]),(930063,array[1,10,50,200,500]::bigint[])
) curve(definition_id,thresholds)
cross join lateral unnest(curve.thresholds) with ordinality valueset(threshold,level)
on conflict(definition_id,level) do update set progress_needed=excluded.progress_needed,reward_amount=excluded.reward_amount,score_points=excluded.score_points;

insert into polls(id,title,headline,summary,start_message,thanks_message,room_id,reward_badge,enabled)
overriding system value values
 (9501,'Room Games Feedback','Tell us what you think','Cuéntanos qué te parecieron los juegos de sala.','Solo tres preguntas rápidas.','¡Gracias por tu opinión!',153,'ACH_GamesPoll1',true)
on conflict(id) do update set title=excluded.title,headline=excluded.headline,summary=excluded.summary,
 start_message=excluded.start_message,thanks_message=excluded.thanks_message,room_id=153,reward_badge=excluded.reward_badge,enabled=true,updated_at=now();

insert into poll_questions(id,poll_id,sort_order,kind,text_ref,category,answer_type,min_selections,options)
overriding system value values
 (9511,9501,1,1,'¿Qué juego de sala probaste?',0,0,1,'[{"value":"banzai","text":"Battle Banzai","type":0},{"value":"freeze","text":"Freeze","type":0},{"value":"football","text":"Football","type":0},{"value":"tag","text":"Tag","type":0}]'),
 (9512,9501,2,2,'¿Qué partes funcionaron?',0,0,1,'[{"value":"timer","text":"Timer","type":0},{"value":"teams","text":"Equipos","type":0},{"value":"score","text":"Puntaje","type":0}]'),
 (9513,9501,3,0,'Cuéntanos cualquier problema que encontraste.',0,0,0,'[]')
on conflict(id) do update set poll_id=excluded.poll_id,sort_order=excluded.sort_order,kind=excluded.kind,text_ref=excluded.text_ref,
 category=excluded.category,answer_type=excluded.answer_type,min_selections=excluded.min_selections,options=excluded.options;

select setval(pg_get_serial_sequence('achievement_definitions','id'),greatest((select max(id) from achievement_definitions),1));
select setval(pg_get_serial_sequence('polls','id'),greatest((select max(id) from polls),1));
select setval(pg_get_serial_sequence('poll_questions','id'),greatest((select max(id) from poll_questions),1));
--rollback delete from poll_answers where poll_id=9501; delete from polls where id=9501; delete from achievement_definitions where id between 930052 and 930063; update achievement_definitions set enabled=false where id between 930044 and 930047;
