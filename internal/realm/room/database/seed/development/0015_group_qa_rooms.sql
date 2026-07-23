--liquibase formatted sql

--changeset pixels:pixels-room-seed-development-0015-group-qa-rooms context:development
insert into rooms (id,owner_player_id,owner_name,name,description,model_name,max_users,score,category_id,trade_mode,staff_picked)
overriding system value values
    (130,1,'demo','GROUPS QA Creator','Eligible room for creation and rollback tests.','model_a',25,0,2,0,false),
    (131,1,'demo','GROUPS QA Open HQ','Open membership, favorite, decorating and group furniture.','model_a',25,0,2,0,false),
    (132,2,'alice','GROUPS QA Requests HQ','Requests, moderation, filters and pagination.','model_a',25,0,2,0,false),
    (133,1,'demo','GROUPS QA Private HQ','Private membership, information and rights denials.','model_a',25,0,2,0,false),
    (134,1,'demo','GROUPS QA Forum HQ','Forum policies, threads, posts and context menu.','model_a',25,0,2,0,false),
    (135,1,'demo','GROUPS QA Removal HQ','Atomic member furniture return and rights revocation.','model_a',25,0,2,0,false)
on conflict(id) do update set owner_player_id=excluded.owner_player_id,owner_name=excluded.owner_name,
    name=excluded.name,description=excluded.description,model_name=excluded.model_name,max_users=excluded.max_users,
    category_id=excluded.category_id,trade_mode=excluded.trade_mode,staff_picked=false,deleted_at=null,updated_at=now();

insert into room_tags(room_id,tag) values
    (130,'qa-groups-creator'),(131,'qa-groups-open'),(132,'qa-groups-requests'),
    (133,'qa-groups-private'),(134,'qa-groups-forum'),(135,'qa-groups-removal')
on conflict do nothing;

select setval(pg_get_serial_sequence('rooms','id'),greatest((select max(id) from rooms),1));
--rollback delete from room_tags where room_id between 130 and 135; delete from rooms where id between 130 and 135;
