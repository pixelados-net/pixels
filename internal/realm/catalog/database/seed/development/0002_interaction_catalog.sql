--liquibase formatted sql

--changeset pixels:pixels-catalog-seed-development-0002-interaction-catalog context:development
--validCheckSum: 9:f372b8ce9241ac962d0b7fcac6c35819
insert into catalog_pages (
    id, parent_id, name, layout, icon_color, icon_image, min_rank,
    order_num, visible, enabled, club_only
)
overriding system value
values
    (5, 1, 'interactions', 'default_3x3', 1, 4, 1, 4, true, true, false),
    (6, 5, 'interactive_basics', 'default_3x3', 1, 5, 1, 1, true, true, false),
    (7, 5, 'room_mechanics', 'default_3x3', 1, 6, 1, 2, true, true, false),
    (8, 5, 'wall_and_surfaces', 'default_3x3', 1, 7, 1, 3, true, true, false)
on conflict do nothing;

insert into catalog_items (
    id, page_id, definition_id, name, cost_credits, cost_points, points_type,
    amount, limited_stack, limited_sells, club_only, order_num, enabled, extra_data
)
overriding system value
values
    (7, 6, 6, 'aquamarine_gate', 2, 0, -1, 1, 0, 0, false, 1, true, '0'),
    (8, 6, 10, 'holodice', 4, 0, -1, 1, 0, 0, false, 2, true, '0'),
    (9, 6, 11, 'mini_bar', 4, 0, -1, 1, 0, 0, false, 3, true, '0'),
    (10, 6, 12, 'cannon', 4, 0, -1, 1, 0, 0, false, 4, true, '0'),
    (11, 6, 14, 'information_terminal', 4, 0, -1, 1, 0, 0, false, 5, true, '0'),
    (12, 6, 24, 'hand_item_tester', 4, 0, -1, 1, 0, 0, false, 6, true, '0'),
    (13, 7, 7, 'aqua_roller', 7, 0, -1, 1, 0, 0, false, 1, true, '0'),
    (14, 7, 8, 'telephone_box', 5, 0, -1, 1, 0, 0, false, 2, true, '0'),
    (15, 7, 13, 'one_way_gate', 4, 0, -1, 1, 0, 0, false, 3, true, '0'),
    (16, 7, 15, 'mute_area', 4, 0, -1, 1, 0, 0, false, 4, true, '0'),
    (17, 7, 16, 'build_area', 4, 0, -1, 1, 0, 0, false, 5, true, '0'),
    (18, 7, 18, 'pyramid', 4, 0, -1, 1, 0, 0, false, 6, true, '0'),
    (19, 8, 9, 'mood_light', 5, 0, -1, 1, 0, 0, false, 1, true, '0'),
    (20, 8, 17, 'wheel_of_fortune', 5, 0, -1, 1, 0, 0, false, 2, true, '0'),
    (21, 8, 19, 'valentines_card', 3, 0, -1, 1, 0, 0, false, 3, true, '0'),
    (22, 8, 20, 'post_it_note', 2, 0, -1, 1, 0, 0, false, 4, true, '0'),
    (23, 8, 21, 'floor_paint', 3, 0, -1, 1, 0, 0, false, 5, true, '0'),
    (24, 8, 22, 'wallpaper', 3, 0, -1, 1, 0, 0, false, 6, true, '0'),
    (25, 8, 23, 'landscape', 3, 0, -1, 1, 0, 0, false, 7, true, '0')
on conflict do nothing;

select setval(pg_get_serial_sequence('catalog_pages', 'id'), greatest((select max(id) from catalog_pages), 1));
select setval(pg_get_serial_sequence('catalog_items', 'id'), greatest((select max(id) from catalog_items), 1));
--rollback delete from catalog_items where id between 7 and 25;
--rollback delete from catalog_pages where id between 5 and 8;
