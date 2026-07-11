--liquibase formatted sql

--changeset pixels:pixels-permission-seed-development-0004-room-settings-nodes context:development
insert into permission_group_nodes (group_id, node, allowed)
values
    (1, 'room.settings.own.manage', true),
    (3, 'room.settings.any.manage', true)
on conflict (group_id, node) do update set allowed=excluded.allowed;

--rollback delete from permission_group_nodes where (group_id=1 and node='room.settings.own.manage') or (group_id=3 and node='room.settings.any.manage');
