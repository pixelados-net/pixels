--liquibase formatted sql

--changeset pixels:pixels-permission-seed-development-0007-trade-nodes context:development
insert into permission_group_nodes(group_id,node,allowed)
values (3,'trade.moderation.lock',true),(3,'marketplace.admin.manage',true)
on conflict(group_id,node) do update set allowed=excluded.allowed;
--rollback delete from permission_group_nodes where group_id=3 and node in ('trade.moderation.lock','marketplace.admin.manage');
