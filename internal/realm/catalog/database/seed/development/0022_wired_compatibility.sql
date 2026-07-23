--liquibase formatted sql

--changeset pixels:pixels-catalog-seed-development-0022-wired-compatibility context:development
insert into catalog_pages(id,parent_id,name,layout,icon_color,icon_image,min_rank,required_node,order_num,visible,enabled,club_only)
overriding system value
values (1006,1000,'wired_compatibility','default_3x3',1,9,1,'room.wired.compatibility.use',6,true,true,false)
on conflict(id) do update set required_node=excluded.required_node,enabled=true,visible=true;

insert into catalog_items(id,page_id,definition_id,name,cost_credits,cost_points,points_type,amount,limited_stack,limited_sells,club_only,order_num,enabled,extra_data)
overriding system value
values
    (300083,1006,900001,'wf_act_alert',3,0,-1,1,0,0,false,1,true,'0'),
    (300084,1006,900002,'wf_act_give_respect',3,0,-1,1,0,0,false,2,true,'0'),
    (300085,1006,900003,'wf_act_give_handitem',3,0,-1,1,0,0,false,3,true,'0'),
    (300086,1006,900004,'wf_act_give_effect',3,0,-1,1,0,0,false,4,true,'0'),
    (300087,1006,900005,'wf_trg_game_team_win',3,0,-1,1,0,0,false,5,true,'0'),
    (300088,1006,900006,'wf_trg_game_team_lose',3,0,-1,1,0,0,false,6,true,'0'),
    (300089,1006,900007,'wf_xtra_or_eval',3,0,-1,1,0,0,false,7,true,'0'),
    (300090,1006,900008,'wf_cnd_valid_moves',3,0,-1,1,0,0,false,8,true,'0')
on conflict(id) do update set definition_id=excluded.definition_id,enabled=true;

select setval(pg_get_serial_sequence('catalog_pages','id'),greatest((select max(id) from catalog_pages),1));
select setval(pg_get_serial_sequence('catalog_items','id'),greatest((select max(id) from catalog_items),1));
--rollback delete from catalog_items where id between 300083 and 300090; delete from catalog_pages where id=1006;
