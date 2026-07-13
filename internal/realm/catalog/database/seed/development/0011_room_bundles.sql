--liquibase formatted sql

--changeset pixels:pixels-catalog-seed-development-0011-room-bundles context:development
insert into catalog_pages (id,parent_id,name,layout,icon_color,icon_image,order_num,visible,enabled,club_only,new_additions)
overriding system value values (107,1,'room_bundles','room_bundle',1,4,7,true,true,false,true)
on conflict (id) do update set parent_id=excluded.parent_id,name=excluded.name,layout=excluded.layout,
    icon_color=excluded.icon_color,icon_image=excluded.icon_image,order_num=excluded.order_num,
    visible=excluded.visible,enabled=excluded.enabled,club_only=excluded.club_only,new_additions=excluded.new_additions;

insert into catalog_items (id,page_id,definition_id,room_bundle_template_room_id,name,cost_credits,cost_points,points_type,amount,limited_stack,limited_sells,bundle_discount_enabled,giftable,club_only,order_num,enabled,extra_data)
overriding system value values
    (1101,107,null,100,'starter_loft_bundle',75,0,-1,0,0,0,false,false,false,1,true,'0'),
    (1102,107,null,101,'interactive_lounge_bundle',120,0,-1,0,0,0,false,false,false,2,true,'0'),
    (1103,107,null,102,'cozy_bedroom_bundle',60,0,-1,0,0,0,false,false,false,3,true,'0')
on conflict (id) do update set page_id=excluded.page_id,definition_id=null,
    room_bundle_template_room_id=excluded.room_bundle_template_room_id,name=excluded.name,
    cost_credits=excluded.cost_credits,cost_points=excluded.cost_points,points_type=excluded.points_type,
    amount=0,limited_stack=0,limited_sells=0,bundle_discount_enabled=false,giftable=false,
    club_only=excluded.club_only,order_num=excluded.order_num,enabled=excluded.enabled,deleted_at=null,updated_at=now();

select setval(pg_get_serial_sequence('catalog_pages','id'),greatest((select max(id) from catalog_pages),1));
select setval(pg_get_serial_sequence('catalog_items','id'),greatest((select max(id) from catalog_items),1));
--rollback delete from catalog_items where id in (1101,1102,1103); delete from catalog_pages where id=107;
