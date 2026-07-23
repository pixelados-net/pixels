--liquibase formatted sql

--changeset pixels:pixels-permission-seed-development-0002-demo-roles context:development
insert into permission_groups (id, name, weight, prefix, prefix_color, parent_group_id)
overriding system value
values (3, 'moderator', 50, 'Mod', '#3498db', 1)
on conflict (id) do update set
    name=excluded.name,
    weight=excluded.weight,
    prefix=excluded.prefix,
    prefix_color=excluded.prefix_color,
    parent_group_id=excluded.parent_group_id;

insert into permission_group_nodes (group_id, node, allowed)
values (3, 'room.doorbell.answer.any', true)
on conflict (group_id, node) do update set allowed=excluded.allowed;

insert into permission_player_groups (player_id, group_id)
select id, 1 from players where id in (1, 2, 3, 4)
on conflict do nothing;

insert into permission_player_groups (player_id, group_id)
select id, 2 from players where id=1
on conflict do nothing;

insert into permission_player_groups (player_id, group_id)
select id, 3 from players where id=2
on conflict do nothing;

select setval(pg_get_serial_sequence('permission_groups', 'id'), greatest((select max(id) from permission_groups), 1));
--rollback delete from permission_player_groups where player_id=2 and group_id=3;
--rollback delete from permission_group_nodes where group_id=3 and node='room.doorbell.answer.any';
--rollback delete from permission_groups where id=3;
