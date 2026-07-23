--liquibase formatted sql

--changeset pixels:pixels-room-seed-development-0013-wired-lab-extras-room context:development
-- Seventh WIRED lab: roller-driven walk triggers, multi-tile footprints, filtered entry and periodic rewards.
insert into rooms (id,owner_player_id,owner_name,name,description,model_name,max_users,score,category_id,trade_mode,staff_picked)
overriding system value
values
    (116,1,'demo','WIRED QA Extras','Roller walk triggers, multi-tile carpet, filtered entry and minute rewards.','model_a',25,0,2,0,false)
on conflict(id) do update set owner_player_id=excluded.owner_player_id,owner_name=excluded.owner_name,
    name=excluded.name,description=excluded.description,model_name=excluded.model_name,max_users=excluded.max_users,
    category_id=excluded.category_id,trade_mode=excluded.trade_mode,staff_picked=excluded.staff_picked;

insert into room_tags(room_id,tag) values (116,'wired-extras') on conflict do nothing;

select setval(pg_get_serial_sequence('rooms','id'),greatest((select max(id) from rooms),1));
--rollback delete from room_tags where room_id = 116; delete from rooms where id = 116;
