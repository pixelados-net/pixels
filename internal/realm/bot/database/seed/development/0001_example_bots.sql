--liquibase formatted sql
--changeset pixels:pixels-bot-seed-development-0001-examples context:development
insert into bots (owner_player_id, behavior_type, name, motto, figure, gender, chat_auto, chat_random)
select id, 'generic', 'Party Bot', 'Yo!', 'hr-100-61.hd-180-1.ch-210-66.lg-270-82.sh-290-80', 'M', true, true
from players where lower(username)='demo' and not exists (select 1 from bots where name='Party Bot');

insert into bots (owner_player_id, behavior_type, name, motto, figure, gender, chat_auto, chat_random)
select id, 'bartender', 'Frank', 'Need a drink?', 'hr-893-45.hd-180-1.ch-255-64.lg-285-64.sh-290-64', 'M', false, false
from players where lower(username)='demo' and not exists (select 1 from bots where name='Frank');

insert into bots (owner_player_id, behavior_type, name, motto, figure, gender, chat_auto, chat_random)
select id, 'visitor_log', 'Connie', 'I remember every visitor.', 'hr-515-33.hd-600-1.ch-635-70.lg-695-82.sh-730-62', 'F', false, false
from players where lower(username)='demo' and not exists (select 1 from bots where name='Connie');

insert into bots (owner_player_id,room_id,behavior_type,name,motto,figure,gender,x,y,z,rotation,can_walk,chat_auto,chat_random)
select players.id,100,'generic','Bundle Host','Welcome to the complete loft.','hr-893-45.hd-180-1.ch-255-64.lg-285-64.sh-290-64','M',5,2,0,2,true,true,false
from players where lower(username)='demo' and exists (select 1 from rooms where id=100 and is_bundle_template)
and not exists (select 1 from bots where name='Bundle Host');

insert into bot_chat_lines (bot_id, order_num, line)
select id, chat_values.order_num, chat_values.line from bots cross join (values (0, 'Yo!'), (1, 'Hello I''m a real party animal!'), (2, 'Hello!')) as chat_values(order_num,line)
where bots.name='Party Bot' on conflict do nothing;

insert into bot_chat_lines (bot_id,order_num,line)
select id,0,'Welcome to %roomname%, %name% is ready to help.' from bots where name='Bundle Host'
on conflict do nothing;

insert into bot_serve_items (keyword, definition_id)
select serve_values.keyword, definitions.id from (values ('tea'),('coffee')) as serve_values(keyword)
cross join lateral (select id from furniture_definitions order by id limit 1) definitions
on conflict (keyword) do nothing;
--rollback delete from bot_serve_items where keyword in ('tea','coffee'); delete from bots where name in ('Party Bot','Frank','Connie','Bundle Host');
