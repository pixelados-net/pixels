--liquibase formatted sql

--changeset pixels:pixels-group-seed-development-0001-wired-qa context:development
insert into social_groups(id,owner_player_id,name,description,home_room_id,state,can_members_decorate,color_a,color_b,badge_code,member_count)
overriding system value
values (1,1,'Wired QA','WIRED social membership compatibility fixture.',110,0,false,1,1,'WIRED_QA',2)
on conflict(id) do update set owner_player_id=excluded.owner_player_id,name=excluded.name,description=excluded.description,
    home_room_id=excluded.home_room_id,state=excluded.state,can_members_decorate=excluded.can_members_decorate,
    color_a=excluded.color_a,color_b=excluded.color_b,badge_code=excluded.badge_code,member_count=excluded.member_count,
    deactivated_at=null,updated_at=now();

insert into social_group_members(group_id,player_id,role)
select 1,id,case when lower(username)='demo' then 0 else 2 end from players where lower(username) in ('demo','alice')
on conflict(group_id,player_id) do update set role=excluded.role,updated_at=now();

insert into room_social_groups(room_id,group_id) values(110,1)
on conflict(room_id) do update set group_id=excluded.group_id,updated_at=now();

select setval(pg_get_serial_sequence('social_groups','id'),greatest((select max(id) from social_groups),1));
--rollback delete from room_social_groups where group_id=1; delete from social_group_members where group_id=1; delete from social_groups where id=1;
