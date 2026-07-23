--liquibase formatted sql

--changeset pixels:pixels-catalog-seed-development-0007-store-final context:development
insert into catalog_pages (id,parent_id,name,layout,icon_color,icon_image,order_num,visible,enabled,club_only,new_additions)
overriding system value values
    (5,1,'bundles','default_3x3',1,4,4,true,true,false,true),
    (6,null,'habbo_club','club_buy',2,5,5,true,true,false,false),
    (7,null,'habbo_club_gifts','club_gift',2,6,6,true,true,true,false)
on conflict (id) do update set name=excluded.name,layout=excluded.layout,new_additions=excluded.new_additions;

insert into catalog_items (id,page_id,definition_id,name,cost_credits,cost_points,points_type,amount,bundle_discount_enabled,giftable,order_num,enabled,extra_data)
overriding system value values
    (7,5,3,'living_set_bundle',45,0,-1,0,false,true,1,true,'0'),
    (8,2,2,'chair_plasto_bulk',2,0,-1,1,true,true,2,true,'0'),
    (9,7,3,'hc_monthly_sofa',0,0,-1,1,false,false,1,true,'0')
on conflict (id) do update set bundle_discount_enabled=excluded.bundle_discount_enabled,giftable=excluded.giftable;

insert into catalog_item_products (catalog_item_id,definition_id,quantity,order_num) values
    (7,3,1,1),(7,1,2,2)
on conflict do nothing;

insert into catalog_vouchers (code,cost_credits,catalog_item_id,redemption_cap,per_player_cap,enabled) values
    ('WELCOME2026',0,1,null,1,true)
on conflict do nothing;

select setval(pg_get_serial_sequence('catalog_pages','id'),greatest((select max(id) from catalog_pages),1));
select setval(pg_get_serial_sequence('catalog_items','id'),greatest((select max(id) from catalog_items),1));
--rollback delete from catalog_voucher_redemptions where voucher_id in (select id from catalog_vouchers where code='WELCOME2026'); delete from catalog_vouchers where code='WELCOME2026'; delete from catalog_item_products where catalog_item_id=7; delete from catalog_items where id in (7,8,9); delete from catalog_pages where id in (5,6,7);
