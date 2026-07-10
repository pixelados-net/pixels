--liquibase formatted sql

--changeset pixels:pixels-permission-seed-test-0001-groups context:test
insert into permission_groups (id, name, weight, prefix, prefix_color, parent_group_id)
overriding system value
values
    (1, 'member', 0, '', '', null),
    (2, 'admin', 100, 'Admin', '#e74c3c', 1)
on conflict do nothing;

insert into permission_group_nodes (group_id, node, allowed)
values (2, '*', true)
on conflict (group_id, node) do update set allowed=excluded.allowed;

insert into permission_player_groups (player_id, group_id)
select id, 1 from players
on conflict do nothing;

select setval(pg_get_serial_sequence('permission_groups', 'id'), greatest((select max(id) from permission_groups), 1));
--rollback delete from permission_player_groups where group_id=1;
--rollback delete from permission_group_nodes where group_id=2 and node='*';
--rollback delete from permission_groups where id in (1, 2);
