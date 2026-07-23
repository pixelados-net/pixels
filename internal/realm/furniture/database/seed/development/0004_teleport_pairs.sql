--liquibase formatted sql

--changeset pixels:pixels-furniture-seed-development-0004-teleport-pairs context:development
insert into furniture_definitions (
    id, sprite_id, name, public_name, kind, width, length, stack_height,
    allow_stack, allow_walk, allow_sit, allow_lay, allow_inventory_stack,
    interaction_type, interaction_modes_count, multiheight, custom_params, metadata
)
overriding system value
values
    (9, 202, 'teleport_tile', 'Teleport Tile', 'floor', 1, 1, 0.00, true, true, false, false, true, 'teleport_tile', 1, '', '', '{}')
on conflict (id) do update set interaction_type = excluded.interaction_type, allow_walk = excluded.allow_walk, updated_at = now();

insert into furniture_items (id, definition_id, owner_player_id, room_id, x, y, z, rotation, extra_data)
overriding system value
values
    (1001, 8, 1, 1, 5, 10, 0, 0, '0'),
    (1002, 8, 1, 1, 10, 10, 0, 0, '0'),
    (1003, 8, 1, 1, 5, 12, 0, 0, '0'),
    (1004, 8, 1, 2, 7, 7, 0, 0, '0'),
    (1005, 9, 1, 1, 6, 12, 0, 0, '0'),
    (1006, 9, 1, 1, 9, 12, 0, 0, '0')
on conflict (id) do update set room_id = excluded.room_id, x = excluded.x, y = excluded.y, z = excluded.z, rotation = excluded.rotation, extra_data = '0', updated_at = now();

insert into furniture_item_teleport_pairs (item_one_id, item_two_id)
values (1001, 1002), (1003, 1004), (1005, 1006)
on conflict (item_one_id) do update set item_two_id = excluded.item_two_id, updated_at = now();
--rollback delete from furniture_item_teleport_pairs where item_one_id in (1001, 1003, 1005);
--rollback delete from furniture_items where id between 1001 and 1006;
--rollback delete from furniture_definitions where id = 9;
