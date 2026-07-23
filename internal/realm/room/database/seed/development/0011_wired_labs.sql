--liquibase formatted sql

--changeset pixels:pixels-room-seed-development-0011-wired-labs context:development
-- Six isolated QA rooms cover avatar/chat, timers, movement, game, bots, and durable rewards.
insert into rooms (id,owner_player_id,owner_name,name,description,model_name,max_users,score,category_id,trade_mode,staff_picked)
overriding system value
values
    (110,1,'demo','WIRED QA Avatar','Stacks 1-2: enter message and keyword toggle.','model_a',25,0,2,0,true),
    (111,1,'demo','WIRED QA Timers','Periodic, random, unseen, delay, reset and call-stack fixtures.','model_a',25,0,2,0,true),
    (112,1,'demo','WIRED QA Movement','Walk, collision, selected furniture and movement effects.','model_a',25,0,2,0,true),
    (113,1,'demo','WIRED QA Game','Teams, blobs and classic/team/wins scoreboards.','model_a',25,0,2,0,true),
    (114,1,'demo','WIRED QA Bots','Bot speech, movement, follow, clothing and hand-item fixtures.','model_a',25,0,2,0,true),
    (115,1,'demo','WIRED QA Rewards','Atomic rewards, room moderation and trace-budget fixtures.','model_a',25,0,2,0,true)
on conflict(id) do update set owner_player_id=excluded.owner_player_id,owner_name=excluded.owner_name,
    name=excluded.name,description=excluded.description,model_name=excluded.model_name,max_users=excluded.max_users,
    category_id=excluded.category_id,trade_mode=excluded.trade_mode,staff_picked=excluded.staff_picked;

insert into room_tags(room_id,tag)
select room_id,tag from (values
    (110::bigint,'wired-avatar'),(111,'wired-timers'),(112,'wired-movement'),
    (113,'wired-game'),(114,'wired-bots'),(115,'wired-rewards')
) values_row(room_id,tag)
on conflict do nothing;

select setval(pg_get_serial_sequence('rooms','id'),greatest((select max(id) from rooms),1));
--rollback delete from room_tags where room_id between 110 and 115; delete from rooms where id between 110 and 115;
