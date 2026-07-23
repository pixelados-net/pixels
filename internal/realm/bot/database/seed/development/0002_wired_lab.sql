--liquibase formatted sql

--changeset pixels:pixels-bot-seed-development-0002-wired-lab context:development
--validCheckSum:ANY
insert into bots(owner_player_id,room_id,behavior_type,name,motto,figure,gender,x,y,z,rotation,can_walk,chat_auto,chat_random)
select id,114,'generic','WiredGuide','WIRED speech and hand-item target.','hr-893-45.hd-180-1.ch-255-64.lg-285-64.sh-290-64','M',5,5,0,2,true,false,false
from players where lower(username)='milo'
and exists(select 1 from rooms where id=114)
and not exists(select 1 from bots where room_id=114 and name='WiredGuide');

insert into bots(owner_player_id,room_id,behavior_type,name,motto,figure,gender,x,y,z,rotation,can_walk,chat_auto,chat_random)
select id,114,'generic','WiredRunner','WIRED movement and follow target.','hr-515-33.hd-600-1.ch-635-70.lg-695-82.sh-730-62','F',9,5,0,6,true,false,false
from players where lower(username)='milo'
and exists(select 1 from rooms where id=114)
and not exists(select 1 from bots where room_id=114 and name='WiredRunner');

insert into bot_chat_lines(bot_id,order_num,line)
select id,0,'Say pixels to exercise the WIRED speech bridge.' from bots where room_id=114 and name='WiredGuide'
on conflict do nothing;
--rollback delete from bots where room_id=114 and name in ('WiredGuide','WiredRunner');
