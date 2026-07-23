--liquibase formatted sql

--changeset pixels:pixels-catalog-seed-development-0004-generic-interaction-examples context:development
--validCheckSum: 9:6d87ddebed1fcca77d84dd527096ca2e
insert into catalog_items (
    id, page_id, definition_id, name, cost_credits, cost_points, points_type,
    amount, limited_stack, limited_sells, club_only, order_num, enabled, extra_data
)
overriding system value
values
    (26, 6, 25, 'flatscreen_tv', 4, 0, -1, 1, 0, 0, false, 7, true, '0'),
    (27, 6, 26, 'table_lamp', 3, 0, -1, 1, 0, 0, false, 8, true, '0'),
    (28, 6, 27, 'color_wheel', 4, 0, -1, 1, 0, 0, false, 9, true, '0')
on conflict (id) do update set
    page_id = excluded.page_id,
    definition_id = excluded.definition_id,
    name = excluded.name,
    cost_credits = excluded.cost_credits,
    order_num = excluded.order_num,
    enabled = excluded.enabled,
    updated_at = now();

select setval(pg_get_serial_sequence('catalog_items', 'id'), greatest((select max(id) from catalog_items), 1));
--rollback delete from catalog_items where id in (26, 27, 28);
