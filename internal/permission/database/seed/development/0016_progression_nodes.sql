--liquibase formatted sql

--changeset pixels:permission-seed-0016-progression-nodes labels:development
insert into permission_group_nodes(group_id,node,allowed)
select id,node,true from permission_groups cross join(values
    ('progression.definitions.manage.any'),
    ('progression.player.override.any'),
    ('progression.quest.manage.any')
) nodes(node) where permission_groups.name='admin'
on conflict(group_id,node) do update set allowed=true;

insert into permission_group_nodes(group_id,node,allowed)
select id,'progression.player.override.any',true from permission_groups where name='moderator'
on conflict(group_id,node) do update set allowed=true;

--rollback delete from permission_group_nodes where node in('progression.definitions.manage.any','progression.player.override.any','progression.quest.manage.any');
