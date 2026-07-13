--liquibase formatted sql

--changeset pixels:pixels-catalog-seed-development-0005-essential-interactions context:development
--validCheckSum: 9:61d9273eb14523109b9fae8b31fc70dd
insert into catalog_items (
    id, page_id, definition_id, name, cost_credits, cost_points, points_type,
    amount, limited_stack, limited_sells, club_only, order_num, enabled, extra_data
)
overriding system value
values
    (29, 6, 28, 'pressure_plate', 4, 0, -1, 1, 0, 0, false, 10, true, '0'),
    (30, 6, 29, 'floor_switch', 4, 0, -1, 1, 0, 0, false, 11, true, '0'),
    (31, 6, 30, 'pink_pura_block', 5, 0, -1, 1, 0, 0, false, 12, true, '0')
on conflict (id) do update set
    page_id = excluded.page_id,
    definition_id = excluded.definition_id,
    name = excluded.name,
    cost_credits = excluded.cost_credits,
    order_num = excluded.order_num,
    enabled = excluded.enabled,
    updated_at = now();

select setval(pg_get_serial_sequence('catalog_items', 'id'), greatest((select max(id) from catalog_items), 1));
--rollback delete from catalog_items where id in (29, 30, 31);
