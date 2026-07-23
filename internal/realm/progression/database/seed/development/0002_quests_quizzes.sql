--liquibase formatted sql
--changeset pixels:pixels-progression-seed-development-0002-quests-quizzes context:development
insert into quest_campaigns(code,seasonal,starts_at,ends_at,timing_code,enabled) values
 ('explore',false,null,null,'',true),('social',false,null,null,'',true),('room_builder',false,null,null,'',true),
 ('seasonal_demo',true,'2020-01-01 00:00:00+00','2035-01-01 00:00:00+00','seasonal_demo',true)
on conflict(code) do update set seasonal=excluded.seasonal,starts_at=excluded.starts_at,ends_at=excluded.ends_at,timing_code=excluded.timing_code,enabled=true;
insert into quest_definitions(id,campaign_code,series_number,name,localization_code,trigger_key,goal_amount,goal_data,reward_kind,reward_currency_type,reward_amount,reward_badge,daily,easy,sort_order,enabled) overriding system value values
 (940001,'explore',1,'enter_rooms','explore.enter_rooms','room.entered',3,'','currency',0,10,'',true,true,1,true),
 (940002,'explore',3,'buy_photo','explore.buy_photo','camera.photo.purchased',1,'','badge',0,0,'ACH_QuestExplore',false,true,3,true),
 (940003,'social',1,'give_respect','social.give_respect','respect.given',2,'','currency',0,10,'',true,true,1,true),
 (940004,'social',2,'make_friend','social.make_friend','friend.request.quest',1,'','badge',0,0,'ACH_QuestSocial',false,true,2,true),
 (940005,'room_builder',1,'place_furniture','room_builder.place','room.furni.count',3,'','currency',0,10,'',true,true,1,true),
 (940006,'room_builder',2,'decorate_floor','room_builder.floor','room.deco.floor',1,'','badge',0,0,'ACH_QuestBuilder',false,true,2,true),
 (940007,'seasonal_demo',1,'seasonal_presence','seasonal.presence','player.presence.minutes',5,'','currency',0,15,'',false,false,1,true),
 (940008,'explore',2,'find_plasto_table','explore.find_plasto_table','room.furni.count',1,'1','currency',0,10,'',false,true,2,true)
on conflict(id) do update set campaign_code=excluded.campaign_code,series_number=excluded.series_number,name=excluded.name,localization_code=excluded.localization_code,trigger_key=excluded.trigger_key,goal_amount=excluded.goal_amount,goal_data=excluded.goal_data,reward_kind=excluded.reward_kind,reward_currency_type=excluded.reward_currency_type,reward_amount=excluded.reward_amount,reward_badge=excluded.reward_badge,daily=excluded.daily,easy=excluded.easy,sort_order=excluded.sort_order,enabled=true;
insert into quizzes(code,kind,enabled) values('SafetyQuiz','safety',true) on conflict(code) do update set kind=excluded.kind,enabled=true;
insert into quiz_questions(id,quiz_code,question_ref,correct_answer_id) overriding system value values
 (950001,'SafetyQuiz',1,2),(950002,'SafetyQuiz',2,1),(950003,'SafetyQuiz',3,3),(950004,'SafetyQuiz',4,2),(950005,'SafetyQuiz',5,1)
on conflict(id) do update set question_ref=excluded.question_ref,correct_answer_id=excluded.correct_answer_id;
insert into promo_badges(code,badge_code,starts_at,ends_at,max_claims,enabled) values
 ('PIXELS2026','PIXELS2026',null,null,0,true)
on conflict(code) do update set badge_code=excluded.badge_code,starts_at=excluded.starts_at,ends_at=excluded.ends_at,max_claims=excluded.max_claims,enabled=true;
select setval(pg_get_serial_sequence('quest_definitions','id'),greatest((select max(id) from quest_definitions),1));
select setval(pg_get_serial_sequence('quiz_questions','id'),greatest((select max(id) from quiz_questions),1));
--rollback delete from promo_badges where code='PIXELS2026'; delete from quizzes where code='SafetyQuiz'; delete from quest_definitions where id between 940001 and 940008; delete from quest_campaigns where code in('explore','social','room_builder','seasonal_demo');
