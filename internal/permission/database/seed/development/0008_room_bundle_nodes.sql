--liquibase formatted sql

--changeset pixels:pixels-permission-seed-development-0008-room-bundle-nodes context:development
insert into permission_group_nodes(group_id,node,allowed)
values (3,'room.admin.bundle_template.manage',true)
on conflict(group_id,node) do update set allowed=excluded.allowed;
--rollback delete from permission_group_nodes where group_id=3 and node='room.admin.bundle_template.manage';
