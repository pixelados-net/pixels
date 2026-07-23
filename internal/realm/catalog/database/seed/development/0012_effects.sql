--liquibase formatted sql

--changeset pixels:pixels-catalog-seed-development-0012-effects context:development
insert into catalog_pages (id,parent_id,name,layout,icon_color,icon_image,order_num,visible,enabled,club_only)
overriding system value values (108,1,'avatar_effects','default_3x3',3,7,8,true,true,false)
on conflict (id) do update set parent_id=excluded.parent_id,name=excluded.name,layout=excluded.layout,visible=true,enabled=true;

insert into catalog_items (id,page_id,definition_id,name,cost_credits,cost_points,points_type,amount,grants_effect_id,grants_effect_duration_seconds,club_only,order_num,enabled,extra_data,giftable)
overriding system value values
    (1201,108,null,'effect_confetti_permanent',15,0,-1,0,101,0,false,1,true,'0',false),
    (1202,108,null,'effect_flames_1day',8,0,-1,0,103,86400,false,2,true,'0',false),
    (1203,108,null,'effect_hc_aura_permanent',0,0,-1,0,201,0,true,3,true,'0',false)
on conflict (id) do update set page_id=excluded.page_id,definition_id=null,grants_effect_id=excluded.grants_effect_id,
    grants_effect_duration_seconds=excluded.grants_effect_duration_seconds,cost_credits=excluded.cost_credits,
    club_only=excluded.club_only,enabled=true,giftable=false,deleted_at=null,updated_at=now();

select setval(pg_get_serial_sequence('catalog_pages','id'),greatest((select max(id) from catalog_pages),1));
select setval(pg_get_serial_sequence('catalog_items','id'),greatest((select max(id) from catalog_items),1));
--rollback delete from catalog_items where id in (1201,1202,1203); delete from catalog_pages where id=108;
