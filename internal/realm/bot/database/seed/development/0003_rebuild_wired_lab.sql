--liquibase formatted sql

--changeset pixels:pixels-bot-seed-development-0003-rebuild-wired-lab context:development
delete from bots where room_id = 114;

insert into bots(owner_player_id,room_id,behavior_type,name,motto,figure,gender,x,y,z,rotation,can_walk,chat_auto,chat_random)
select id,114,'generic','WiredGuide','Public talk, private talk, clothing and hand-item WIRED target.',
    'hr-893-45.hd-180-1.ch-255-64.lg-285-64.sh-290-64','M',5,5,0,2,true,false,false
from players where lower(username)='demo';

insert into bots(owner_player_id,room_id,behavior_type,name,motto,figure,gender,x,y,z,rotation,can_walk,chat_auto,chat_random)
select id,114,'generic','WiredRunner','Movement, teleport, follow and arrival-trigger WIRED target.',
    'hr-515-33.hd-600-1.ch-635-70.lg-695-82.sh-730-62','F',9,5,0,6,true,false,false
from players where lower(username)='demo';

--rollback delete from bots where room_id=114 and name in ('WiredGuide','WiredRunner');
