--liquibase formatted sql
--changeset pixels:permission-seed-0014-crafting-nodes labels:development
insert into permission_group_nodes(group_id,node,allowed)
select id,node,true from permission_groups cross join(values('crafting.altar.manage.any'),('crafting.recycler.manage.any'),('crafting.player.override.any')) nodes(node)
where permission_groups.name='admin' on conflict(group_id,node) do update set allowed=true;
insert into permission_group_nodes(group_id,node,allowed)
select id,node,true from permission_groups cross join(values('crafting.player.override.any')) nodes(node)
where permission_groups.name='moderator' on conflict(group_id,node) do update set allowed=true;
--rollback delete from permission_group_nodes where node in('crafting.altar.manage.any','crafting.recycler.manage.any','crafting.player.override.any');
