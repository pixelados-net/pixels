--liquibase formatted sql

--changeset pixels:pixels-room-seed-development-0015-group-qa-rooms context:development
--validCheckSum:ANY
insert into rooms (id,owner_player_id,owner_name,name,description,model_name,max_users,score,category_id,trade_mode,staff_picked)
overriding system value values
    (130,1,'milo','Founders'' Hall','Thinking of starting a group? This is where it begins.','model_a',25,0,2,0,false),
    (131,1,'milo','The Open Circle HQ','Open membership - anyone can join and help decorate.','model_a',25,0,2,0,false),
    (132,2,'juno','Skyline Requests HQ','Membership by request - tell us why you would like to join.','model_a',25,0,2,0,false),
    (133,1,'milo','The Hidden Society HQ','Invite-only. Members keep this space exclusive.','model_a',25,0,2,0,false),
    (134,1,'milo','Town Hall','Our community forum HQ - announcements and discussion.','model_a',25,0,2,0,false),
    (135,1,'milo','Sunset Terrace','A quiet spot to wind down after a long day.','model_a',25,0,2,0,false)
on conflict(id) do update set owner_player_id=excluded.owner_player_id,owner_name=excluded.owner_name,
    name=excluded.name,description=excluded.description,model_name=excluded.model_name,max_users=excluded.max_users,
    category_id=excluded.category_id,trade_mode=excluded.trade_mode,staff_picked=false,deleted_at=null,updated_at=now();

insert into room_tags(room_id,tag) values
    (130,'groups'),(131,'open'),(132,'requests'),
    (133,'private'),(134,'forum'),(135,'social')
on conflict do nothing;

select setval(pg_get_serial_sequence('rooms','id'),greatest((select max(id) from rooms),1));
--rollback delete from room_tags where room_id between 130 and 135; delete from rooms where id between 130 and 135;
