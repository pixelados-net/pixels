--liquibase formatted sql

--changeset pixels:permission-seed-0013-group-nodes labels:development
insert into permission_group_nodes(group_id,node,allowed)
select id,node,true
from permission_groups
cross join (values
    ('group.create')
) as nodes(node)
where permission_groups.name='member'
on conflict (group_id,node) do update set allowed=excluded.allowed;

insert into permission_group_nodes(group_id,node,allowed)
select id,node,true
from permission_groups
cross join (values
    ('group.create'),
    ('group.manage.any'),
    ('group.delete.any'),
    ('group.members.manage.any'),
    ('group.roles.manage.any'),
    ('group.home_room.rebind'),
    ('group.badge.manage.any'),
    ('group.forum.manage.any'),
    ('group.forum.moderate.any'),
    ('group.read.deactivated')
) as nodes(node)
where permission_groups.name='admin'
on conflict (group_id,node) do update set allowed=excluded.allowed;

insert into permission_group_nodes(group_id,node,allowed)
select id,node,true
from permission_groups
cross join (values
    ('group.members.manage.any'),
    ('group.forum.moderate.any')
) as nodes(node)
where permission_groups.name='moderator'
on conflict (group_id,node) do update set allowed=excluded.allowed;

--rollback delete from permission_group_nodes where node in ('group.create','group.manage.any','group.delete.any','group.members.manage.any','group.roles.manage.any','group.home_room.rebind','group.badge.manage.any','group.forum.manage.any','group.forum.moderate.any','group.read.deactivated');
