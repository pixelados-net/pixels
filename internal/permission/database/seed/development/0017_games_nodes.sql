--liquibase formatted sql

--changeset pixels:permission-games-0017 context:development
insert into permission_group_nodes(group_id,node,allowed)
select permission_group.id, value.node, true
from permission_groups permission_group
cross join (values ('games.center.manage.any'),('games.polls.manage.any')) value(node)
where permission_group.name='admin'
on conflict(group_id,node) do update set allowed=true;

insert into permission_group_nodes(group_id,node,allowed)
select permission_group.id,'games.polls.manage.any',true
from permission_groups permission_group
where permission_group.name='moderator'
on conflict(group_id,node) do update set allowed=true;

--rollback delete from permission_group_nodes where node in ('games.center.manage.any','games.polls.manage.any');
