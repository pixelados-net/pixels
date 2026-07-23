--liquibase formatted sql

--changeset pixels:pixels-permission-seed-development-0009-player-effect-nodes context:development
insert into permission_group_nodes(group_id,node,allowed)
select id,'player.admin.effect.grant',true from permission_groups where name='admin'
on conflict(group_id,node) do update set allowed=excluded.allowed;

--rollback delete from permission_group_nodes where node='player.admin.effect.grant';
