--liquibase formatted sql

--changeset pixels:pixels-permission-seed-development-0010-moderation-nodes context:development
--validCheckSum:ANY
insert into permission_group_nodes(group_id,node,allowed)
select id,node,true
from permission_groups
cross join (values
    ('moderation.tool.access'),
    ('moderation.issue.manage'),
    ('moderation.chatlog.read'),
    ('moderation.sanction.apply'),
    ('moderation.room.override'),
    ('moderation.guide.duty'),
    ('moderation.guardian.duty')
) as nodes(node)
where name='moderator'
on conflict(group_id,node) do update set allowed=excluded.allowed;

insert into permission_player_nodes(player_id,node,allowed)
select id,'moderation.guide.duty',true from players where lower(username)='reid'
on conflict(player_id,node) do update set allowed=excluded.allowed;

insert into permission_player_nodes(player_id,node,allowed)
select id,'moderation.guardian.duty',true from players where lower(username)='wren'
on conflict(player_id,node) do update set allowed=excluded.allowed;

--rollback delete from permission_player_nodes where node in ('moderation.guide.duty','moderation.guardian.duty'); delete from permission_group_nodes where node like 'moderation.%';
