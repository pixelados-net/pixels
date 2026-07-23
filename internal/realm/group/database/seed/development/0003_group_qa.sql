--liquibase formatted sql

--changeset pixels:pixels-group-seed-development-0003-group-qa context:development
insert into social_groups(id,owner_player_id,name,description,home_room_id,state,can_members_decorate,color_a,color_b,badge_code,forum_enabled,member_count,pending_count,thread_count,post_count)
overriding system value values
    (2,1,'Pixels Open','Open social group QA fixture.',131,0,true,1,2,'b001010s004020',false,3,0,0,0),
    (3,2,'Pixels Requests','Exclusive request and role QA fixture.',132,1,false,3,4,'b002030s005040',false,3,1,0,0),
    (4,1,'Pixels Private','Private membership and rights QA fixture.',133,2,false,5,6,'b003050s006060',false,2,0,0,0),
    (5,1,'Pixels Forum QA','Forum policies and moderation QA fixture.',134,0,true,7,8,'b004070s007080',true,3,0,1,2)
on conflict(id) do update set owner_player_id=excluded.owner_player_id,name=excluded.name,description=excluded.description,
    home_room_id=excluded.home_room_id,state=excluded.state,can_members_decorate=excluded.can_members_decorate,
    color_a=excluded.color_a,color_b=excluded.color_b,badge_code=excluded.badge_code,forum_enabled=excluded.forum_enabled,
    member_count=excluded.member_count,pending_count=excluded.pending_count,thread_count=excluded.thread_count,
    post_count=excluded.post_count,deactivated_at=null,updated_at=now();

insert into social_group_badge_parts(group_id,ordinal,kind,element_id,color_family,color_id,position) values
    (2,0,0,1,0,1,0),(2,1,1,4,1,2,0),
    (3,0,0,2,0,3,0),(3,1,1,5,1,4,0),
    (4,0,0,3,0,5,0),(4,1,1,6,1,6,0),
    (5,0,0,4,0,7,0),(5,1,1,7,1,8,0)
on conflict(group_id,ordinal) do update set kind=excluded.kind,element_id=excluded.element_id,
    color_family=excluded.color_family,color_id=excluded.color_id,position=excluded.position;

insert into social_group_members(group_id,player_id,role) values
    (2,1,0),(2,2,1),(2,3,2),
    (3,2,0),(3,1,1),(3,4,2),
    (4,1,0),(4,2,2),
    (5,1,0),(5,2,1),(5,3,2)
on conflict(group_id,player_id) do update set role=excluded.role,updated_at=now();

insert into social_group_requests(group_id,player_id) values (3,3)
on conflict do nothing;

insert into player_social_group_preferences(player_id,favorite_group_id) values
    (1,2),(2,3),(3,2),(4,3)
on conflict(player_id) do update set favorite_group_id=excluded.favorite_group_id,updated_at=now();

insert into room_social_groups(room_id,group_id) values (131,2),(132,3),(133,4),(134,5)
on conflict(room_id) do update set group_id=excluded.group_id,updated_at=now();

insert into social_group_forum_threads(id,group_id,author_player_id,author_name,subject,post_count,last_post_id,last_author_player_id,last_author_name,last_posted_at)
overriding system value values (5001,5,1,'demo','Welcome to Pixels Forum QA',2,5002,3,'bob',now())
on conflict(id) do update set subject=excluded.subject,post_count=excluded.post_count,last_post_id=excluded.last_post_id,
    last_author_player_id=excluded.last_author_player_id,last_author_name=excluded.last_author_name,last_posted_at=excluded.last_posted_at,updated_at=now();

insert into social_group_forum_posts(id,group_id,thread_id,ordinal,author_player_id,author_name,author_figure,body)
overriding system value values
    (5001,5,5001,0,1,'demo','hr-100.hd-180-1.ch-210-66.lg-270-82.sh-290-80','Use this thread to verify forum read markers and policies.'),
    (5002,5,5001,1,3,'bob','hr-828-61.hd-180-8.ch-255-81.lg-280-64.sh-305-62','Forum reply fixture for unread and moderation tests.')
on conflict(id) do update set body=excluded.body,updated_at=now();

select setval(pg_get_serial_sequence('social_groups','id'),greatest((select max(id) from social_groups),1));
select setval(pg_get_serial_sequence('social_group_forum_threads','id'),greatest((select max(id) from social_group_forum_threads),1));
select setval(pg_get_serial_sequence('social_group_forum_posts','id'),greatest((select max(id) from social_group_forum_posts),1));
--rollback delete from social_group_forum_posts where id between 5001 and 5002; delete from social_group_forum_threads where id=5001; delete from player_social_group_preferences where player_id in (1,2,3,4) and favorite_group_id between 2 and 5; delete from social_group_requests where group_id between 2 and 5; delete from room_social_groups where group_id between 2 and 5; delete from social_group_members where group_id between 2 and 5; delete from social_group_badge_parts where group_id between 2 and 5; delete from social_groups where id between 2 and 5;
