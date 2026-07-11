--liquibase formatted sql

--changeset pixels:pixels-permission-seed-development-0005-room-moderation-policy-nodes context:development
insert into permission_group_nodes (group_id, node, allowed)
values
    (1, 'room.moderation.policy.own.manage', true),
    (3, 'room.moderation.policy.any.manage', true)
on conflict (group_id, node) do update set allowed=excluded.allowed;

--rollback delete from permission_group_nodes where (group_id=1 and node='room.moderation.policy.own.manage') or (group_id=3 and node='room.moderation.policy.any.manage');
