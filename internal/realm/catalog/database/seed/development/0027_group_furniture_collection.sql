--liquibase formatted sql

--changeset pixels:pixels-catalog-seed-development-0027-group-furniture-collection context:development
insert into catalog_items(id,page_id,definition_id,name,cost_credits,cost_points,points_type,amount,limited_stack,limited_sells,bundle_discount_enabled,giftable,club_only,order_num,enabled,extra_data,reward_kind)
overriding system value values
    (1403,116,4249,'group_banner',4,0,-1,1,0,0,false,false,false,2,true,'0','furniture'),
    (1404,116,4248,'group_carpet',3,0,-1,1,0,0,false,false,false,3,true,'0','furniture'),
    (1406,116,5039,'group_flag_1',5,0,-1,1,0,0,false,false,false,5,true,'0','furniture'),
    (1407,116,5037,'group_flag_2',5,0,-1,1,0,0,false,false,false,6,true,'0','furniture'),
    (1408,116,4253,'group_pennant',4,0,-1,1,0,0,false,false,false,7,true,'0','furniture'),
    (1409,116,4254,'group_sofa',3,0,-1,1,0,0,false,false,false,8,true,'0','furniture'),
    (1410,116,4251,'group_stool_1',1,0,-1,1,0,0,false,false,false,9,true,'0','furniture'),
    (1411,116,4250,'group_stool_2',1,0,-1,1,0,0,false,false,false,10,true,'0','furniture'),
    (1412,116,5040,'group_tile_1',5,0,-1,1,0,0,false,false,false,11,true,'0','furniture'),
    (1413,116,5038,'group_tile_2',3,0,-1,1,0,0,false,false,false,12,true,'0','furniture'),
    (1414,116,5041,'group_table',5,0,-1,1,0,0,false,false,false,13,true,'0','furniture'),
    (1415,116,4252,'group_divider',1,0,-1,1,0,0,false,false,false,14,true,'0','furniture')
on conflict(id) do update set page_id=excluded.page_id,definition_id=excluded.definition_id,name=excluded.name,
    cost_credits=excluded.cost_credits,cost_points=excluded.cost_points,points_type=excluded.points_type,
    amount=1,limited_stack=0,limited_sells=0,bundle_discount_enabled=false,giftable=false,
    club_only=false,order_num=excluded.order_num,enabled=excluded.enabled,extra_data=excluded.extra_data,
    reward_kind=excluded.reward_kind,deleted_at=null,updated_at=now();

select setval(pg_get_serial_sequence('catalog_items','id'),greatest((select max(id) from catalog_items),1));

--rollback delete from catalog_items where id between 1403 and 1415;
