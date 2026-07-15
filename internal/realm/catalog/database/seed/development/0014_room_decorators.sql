--liquibase formatted sql

--changeset pixels:pixels-catalog-seed-development-0014-room-decorators context:development
insert into catalog_pages(id,parent_id,name,layout,icon_color,icon_image,min_rank,order_num,visible,enabled,club_only)
overriding system value values
    (900,5,'room_surfaces','spaces_new',1,7,1,4,true,true,false),
    (901,5,'room_decorators','default_3x3',1,7,1,5,true,true,false)
on conflict(id) do update set name=excluded.name,layout=excluded.layout,visible=true,enabled=true;

insert into catalog_items(id,page_id,definition_id,name,cost_credits,cost_points,points_type,amount,limited_stack,limited_sells,club_only,order_num,enabled,extra_data)
overriding system value values
    (9001,900,21,'floor_white',3,0,-1,1,0,0,false,1,true,'101'),
    (9002,900,21,'floor_blue',3,0,-1,1,0,0,false,2,true,'110'),
    (9003,900,21,'floor_dark',3,0,-1,1,0,0,false,3,true,'201'),
    (9004,900,22,'wallpaper_classic',3,0,-1,1,0,0,false,4,true,'101'),
    (9005,900,22,'wallpaper_blue',3,0,-1,1,0,0,false,5,true,'110'),
    (9006,900,22,'wallpaper_dark',3,0,-1,1,0,0,false,6,true,'201'),
    (9007,900,23,'landscape_city',3,0,-1,1,0,0,false,7,true,'1.1'),
    (9008,900,23,'landscape_sunset',3,0,-1,1,0,0,false,8,true,'2.1'),
    (9009,900,23,'landscape_night',3,0,-1,1,0,0,false,9,true,'3.1'),
    (9010,901,40,'sticky_pole',4,0,-1,1,0,0,false,1,true,'0'),
    (9011,901,41,'mannequin',4,0,-1,1,0,0,false,2,true,'{"gender":"M","figure":"hd-99999-99998.ch-210-66.lg-270-82.sh-290-80","name":"My look"}'),
    (9012,901,42,'background_toner',5,0,-1,1,0,0,false,3,true,'0:0:0:0')
on conflict(id) do update set page_id=excluded.page_id,definition_id=excluded.definition_id,name=excluded.name,enabled=true,extra_data=excluded.extra_data;
select setval(pg_get_serial_sequence('catalog_pages','id'),greatest((select max(id) from catalog_pages),1));
select setval(pg_get_serial_sequence('catalog_items','id'),greatest((select max(id) from catalog_items),1));
--rollback delete from catalog_items where id between 9001 and 9012; delete from catalog_pages where id in (900,901);
