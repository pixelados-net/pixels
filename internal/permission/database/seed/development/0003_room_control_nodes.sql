--liquibase formatted sql

--changeset pixels:pixels-permission-seed-development-0003-room-control-nodes context:development
insert into permission_group_nodes (group_id, node, allowed)
values
    (1, 'room.moderation.own.kick', true),
    (1, 'room.moderation.own.mute', true),
    (1, 'room.moderation.own.ban', true),
    (1, 'room.rights.own.grant', true),
    (1, 'room.rights.own.revoke', true),
    (3, 'room.moderation.any.kick', true),
    (3, 'room.moderation.any.mute', true),
    (3, 'room.moderation.any.ban', true),
    (3, 'room.rights.any.grant', true),
    (3, 'room.rights.any.revoke', true)
on conflict (group_id, node) do update set allowed=excluded.allowed;

--rollback delete from permission_group_nodes where (group_id=1 and node in ('room.moderation.own.kick', 'room.moderation.own.mute', 'room.moderation.own.ban', 'room.rights.own.grant', 'room.rights.own.revoke')) or (group_id=3 and node in ('room.moderation.any.kick', 'room.moderation.any.mute', 'room.moderation.any.ban', 'room.rights.any.grant', 'room.rights.any.revoke'));
