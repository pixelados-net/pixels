--liquibase formatted sql

--changeset pixels:pixels-permission-seed-development-0011-wired-nodes context:development
insert into permission_group_nodes(group_id,node,allowed)
select id,node,true
from permission_groups
cross join (values
    ('room.wired.configure'),
    ('room.wired.inspect')
) as nodes(node)
where name='member'
on conflict(group_id,node) do update set allowed=excluded.allowed;

insert into permission_group_nodes(group_id,node,allowed)
select id,node,true
from permission_groups
cross join (values
    ('room.wired.configure.any'),
    ('room.wired.inspect'),
    ('room.wired.admin'),
    ('room.wired.reward.manage'),
    ('room.wired.compatibility.use')
) as nodes(node)
where name='moderator'
on conflict(group_id,node) do update set allowed=excluded.allowed;
--rollback delete from permission_group_nodes where node like 'room.wired.%';
