--liquibase formatted sql

--changeset pixels:pixels-catalog-seed-development-0008-hc-catalog context:development
update catalog_pages set
    parent_id = 1, name = 'interactions', layout = 'default_3x3', icon_color = 1,
    icon_image = 4, order_num = 4, visible = true, enabled = true,
    club_only = false, new_additions = false
where id = 5;

update catalog_pages set
    parent_id = 5, name = 'interactive_basics', layout = 'default_3x3', icon_color = 1,
    icon_image = 5, order_num = 1, visible = true, enabled = true,
    club_only = false, new_additions = false
where id = 6;

update catalog_pages set
    parent_id = 5, name = 'room_mechanics', layout = 'default_3x3', icon_color = 1,
    icon_image = 6, order_num = 2, visible = true, enabled = true,
    club_only = false, new_additions = false
where id = 7;

update catalog_pages set
    parent_id = 5, name = 'wall_and_surfaces', layout = 'default_3x3', icon_color = 1,
    icon_image = 7, order_num = 3, visible = true, enabled = true,
    club_only = false, new_additions = false
where id = 8;

update catalog_items set bundle_discount_enabled = false, giftable = false where id in (7, 8, 9);

insert into catalog_pages (
    id, parent_id, name, layout, icon_color, icon_image, order_num,
    visible, enabled, club_only, new_additions
)
overriding system value values
    (100, null, 'hc', 'default_3x3', 2, 5, 2, true, true, false, false),
    (101, 100, 'habbo_club', 'vip_buy', 2, 5, 1, true, true, false, false),
    (102, 100, 'habbo_club_gifts', 'club_gifts', 2, 6, 2, true, true, false, false),
    (103, 1, 'bundles', 'default_3x3', 1, 4, 5, true, true, false, true)
on conflict (id) do update set
    parent_id = excluded.parent_id, name = excluded.name, layout = excluded.layout,
    icon_color = excluded.icon_color, icon_image = excluded.icon_image,
    order_num = excluded.order_num, visible = excluded.visible, enabled = excluded.enabled,
    club_only = excluded.club_only, new_additions = excluded.new_additions;

insert into catalog_items (
    id, page_id, definition_id, name, cost_credits, cost_points, points_type,
    amount, bundle_discount_enabled, giftable, order_num, enabled, extra_data
)
overriding system value values
    (1001, 103, 3, 'living_set_bundle', 45, 0, -1, 0, false, true, 1, true, '0'),
    (1002, 2, 2, 'chair_plasto_bulk', 2, 0, -1, 1, true, true, 2, true, '0'),
    (1003, 102, 3, 'hc_monthly_sofa', 0, 0, -1, 1, false, false, 1, true, '0')
on conflict (id) do update set
    page_id = excluded.page_id, definition_id = excluded.definition_id,
    name = excluded.name, cost_credits = excluded.cost_credits,
    cost_points = excluded.cost_points, points_type = excluded.points_type,
    amount = excluded.amount, bundle_discount_enabled = excluded.bundle_discount_enabled,
    giftable = excluded.giftable, order_num = excluded.order_num,
    enabled = excluded.enabled, extra_data = excluded.extra_data;

delete from catalog_item_products where catalog_item_id = 1001;
insert into catalog_item_products (catalog_item_id, definition_id, quantity, order_num) values
    (1001, 3, 1, 1),
    (1001, 1, 2, 2);

select setval(pg_get_serial_sequence('catalog_pages','id'), greatest((select max(id) from catalog_pages), 1));
select setval(pg_get_serial_sequence('catalog_items','id'), greatest((select max(id) from catalog_items), 1));
--rollback delete from catalog_item_products where catalog_item_id = 1001; delete from catalog_items where id in (1001, 1002, 1003); delete from catalog_pages where id in (100, 101, 102, 103);
