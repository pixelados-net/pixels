--liquibase formatted sql

--changeset pixels:pixels-furniture-seed-development-0037-group-furniture-collection context:development
insert into furniture_definitions(id,sprite_id,name,public_name,kind,width,length,stack_height,allow_stack,allow_walk,allow_sit,allow_lay,allow_inventory_stack,allow_trade,allow_marketplace_sale,interaction_type,interaction_modes_count,multiheight,custom_params,metadata)
overriding system value values
    (4248,4248,'gld_carpet','Group Carpet','floor',2,3,0.01,true,true,false,false,true,true,false,'group_furniture',11,'','','{"socialGroup":true,"source":"arcturus"}'),
    (4249,4249,'gld_badgewall_tall','Group Banner','floor',1,1,1.00,true,false,false,false,true,true,false,'group_furniture',11,'','','{"socialGroup":true,"source":"arcturus"}'),
    (4250,4250,'gld_stool2','Group Stool II','floor',1,1,1.60,true,false,true,false,true,true,false,'group_furniture',1,'','','{"socialGroup":true,"source":"arcturus"}'),
    (4251,4251,'gld_stool1','Group Stool I','floor',1,1,1.60,true,false,true,false,true,true,false,'group_furniture',1,'','','{"socialGroup":true,"source":"arcturus"}'),
    (4252,4252,'gld_wall_tall','Group Divider','floor',1,1,1.00,true,false,false,false,true,true,false,'group_furniture',11,'','','{"socialGroup":true,"source":"arcturus"}'),
    (4253,4253,'gld_pennant','Group Pennant','floor',1,1,1.00,true,false,false,false,true,true,false,'group_furniture',11,'','','{"socialGroup":true,"source":"arcturus"}'),
    (4254,4254,'gld_sofa1','Group Sofa','floor',2,1,1.10,true,false,true,false,true,true,false,'group_furniture',1,'','','{"socialGroup":true,"source":"arcturus"}'),
    (5037,5037,'gld_hangflag2','Group Flag II','floor',1,1,0.00,true,true,false,false,true,true,false,'group_furniture',1,'','','{"socialGroup":true,"source":"arcturus"}'),
    (5038,5038,'gld_tile2','Group Tile II','floor',2,2,0.00,true,true,false,false,true,true,false,'group_furniture',4,'','','{"socialGroup":true,"source":"arcturus"}'),
    (5039,5039,'gld_hangflag1','Group Flag I','floor',1,1,0.00,true,true,false,false,true,true,false,'group_furniture',1,'','','{"socialGroup":true,"source":"arcturus"}'),
    (5040,5040,'gld_tile1','Group Tile I','floor',2,2,0.00,true,true,false,false,true,true,false,'group_furniture',4,'','','{"socialGroup":true,"source":"arcturus"}'),
    (5041,5041,'gld_table1','Group Table','floor',2,2,1.00,true,true,false,false,true,true,false,'group_furniture',1,'','','{"socialGroup":true,"source":"arcturus"}')
on conflict(id) do update set sprite_id=excluded.sprite_id,name=excluded.name,public_name=excluded.public_name,
    kind=excluded.kind,width=excluded.width,length=excluded.length,stack_height=excluded.stack_height,
    allow_stack=excluded.allow_stack,allow_walk=excluded.allow_walk,allow_sit=excluded.allow_sit,
    allow_lay=excluded.allow_lay,allow_inventory_stack=excluded.allow_inventory_stack,
    allow_trade=excluded.allow_trade,allow_marketplace_sale=excluded.allow_marketplace_sale,
    interaction_type=excluded.interaction_type,interaction_modes_count=excluded.interaction_modes_count,
    metadata=excluded.metadata,deleted_at=null,updated_at=now();

select setval(pg_get_serial_sequence('furniture_definitions','id'),greatest((select max(id) from furniture_definitions),1));

--rollback delete from furniture_definitions where id in (4248,4249,4250,4251,4252,4253,4254,5037,5038,5039,5040,5041);
