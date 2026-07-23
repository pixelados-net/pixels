--liquibase formatted sql

--changeset pixels:pixels-catalog-seed-development-0001-default-catalog context:development
--validCheckSum: 9:94ec78c411280e8b3ba709106899a37f
insert into catalog_pages (id, parent_id, name, layout, icon_color, icon_image, min_rank, order_num, visible, enabled, club_only)
overriding system value
values
    (1, null, 'furniture', 'default_3x3', 1, 1, 1, 1, true, true, false),
    (2, 1, 'chairs', 'default_3x3', 1, 1, 1, 1, true, true, false),
    (3, 1, 'tables', 'default_3x3', 1, 2, 1, 2, true, true, false),
    (4, 1, 'sofas_beds', 'default_3x3', 1, 3, 1, 3, true, true, false)
on conflict do nothing;

insert into catalog_items (id, page_id, definition_id, name, cost_credits, cost_points, points_type, amount, limited_stack, limited_sells, club_only, order_num, enabled, extra_data)
overriding system value
values
    (1, 2, 2, 'chair_plasto', 2, 0, -1, 1, 0, 0, false, 1, true, '0'),
    (2, 3, 1, 'table_plasto_4leg', 15, 0, -1, 1, 0, 0, false, 1, true, '0'),
    (3, 4, 3, 'sofa_silo', 12, 0, -1, 1, 0, 0, false, 1, true, '0'),
    (4, 4, 3, 'sofa_silo_ltd', 0, 15, 5, 1, 10, 0, false, 2, true, '0'),
    (5, 4, 4, 'bed_silo_one', 20, 0, -1, 1, 0, 0, false, 3, true, '0'),
    (6, 4, 5, 'bed_silo_two', 35, 0, -1, 1, 0, 0, false, 4, true, '0')
on conflict do nothing;

insert into catalog_item_limited_units (catalog_item_id, unit_number)
select 4, value from generate_series(1, 10) as value
on conflict do nothing;

select setval(pg_get_serial_sequence('catalog_pages', 'id'), greatest((select max(id) from catalog_pages), 1));
select setval(pg_get_serial_sequence('catalog_items', 'id'), greatest((select max(id) from catalog_items), 1));
--rollback delete from catalog_item_limited_units where catalog_item_id = 4;
--rollback delete from catalog_items where id between 1 and 6;
--rollback delete from catalog_pages where id between 1 and 4;
