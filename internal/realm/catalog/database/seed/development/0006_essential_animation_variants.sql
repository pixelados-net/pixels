--liquibase formatted sql

--changeset pixels:pixels-catalog-seed-development-0006-essential-animation-variants context:development
--validCheckSum: 9:7608fc7e162960b14670ca753ea237b2
insert into catalog_items (
    id, page_id, definition_id, name, cost_credits, cost_points, points_type,
    amount, limited_stack, limited_sells, club_only, order_num, enabled, extra_data
)
overriding system value
values
    (32, 6, 31, 'magic_crystal_random', 4, 0, -1, 1, 0, 0, false, 13, true, '0'),
    (33, 6, 32, 'arrow_color_plate', 4, 0, -1, 1, 0, 0, false, 14, true, '0'),
    (34, 6, 33, 'remote_floor_switch', 4, 0, -1, 1, 0, 0, false, 15, true, '0'),
    (35, 6, 34, 'pura_fridge_no_sides', 4, 0, -1, 1, 0, 0, false, 16, true, '0'),
    (36, 6, 35, 'sink_hand_item', 3, 0, -1, 1, 0, 0, false, 17, true, '0')
on conflict (id) do update set
    page_id = excluded.page_id,
    definition_id = excluded.definition_id,
    name = excluded.name,
    cost_credits = excluded.cost_credits,
    order_num = excluded.order_num,
    enabled = excluded.enabled,
    updated_at = now();

select setval(pg_get_serial_sequence('catalog_items', 'id'), greatest((select max(id) from catalog_items), 1));
--rollback delete from catalog_items where id in (32, 33, 34, 35, 36);
