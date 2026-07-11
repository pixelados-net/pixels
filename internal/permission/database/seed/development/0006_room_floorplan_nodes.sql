--liquibase formatted sql

--changeset pixels:pixels-permission-seed-0006-room-floorplan-nodes context:development
insert into permission_group_nodes (group_id, node, allowed)
values
    (1, 'room.floorplan.own.edit', true),
    (3, 'room.floorplan.any.edit', true)
on conflict (group_id, node) do update set allowed = excluded.allowed;

--rollback delete from permission_group_nodes where (group_id=1 and node='room.floorplan.own.edit') or (group_id=3 and node='room.floorplan.any.edit');
