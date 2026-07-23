--liquibase formatted sql

--changeset pixels:pixels-catalog-seed-development-0010-marketplace-pages context:development
insert into catalog_pages (
    id, parent_id, name, layout, icon_color, icon_image, order_num,
    visible, enabled, club_only, new_additions
)
overriding system value values
    (104, 1, 'marketplace', 'default_3x3', 1, 8, 6, true, true, false, false),
    (105, 104, 'marketplace_own_items', 'marketplace_own_items', 1, 9, 1, true, true, false, false),
    (106, 104, 'marketplace_offers', 'marketplace', 1, 10, 2, true, true, false, false)
on conflict (id) do update set
    parent_id = excluded.parent_id, name = excluded.name, layout = excluded.layout,
    icon_color = excluded.icon_color, icon_image = excluded.icon_image,
    order_num = excluded.order_num, visible = excluded.visible, enabled = excluded.enabled,
    club_only = excluded.club_only, new_additions = excluded.new_additions;

select setval(pg_get_serial_sequence('catalog_pages','id'), greatest((select max(id) from catalog_pages), 1));
--rollback delete from catalog_pages where id in (105, 106, 104);
