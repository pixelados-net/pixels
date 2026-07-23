--liquibase formatted sql

--changeset pixels:pixels-furniture-seed-development-0035-monsterplant-mappings context:development
-- Definitions use the current official Habbo floor-furnidata identifiers and class names.
insert into furniture_definitions (id,sprite_id,name,public_name,kind,width,length,stack_height,allow_stack,allow_walk,allow_sit,allow_lay,allow_inventory_stack,allow_trade,allow_marketplace_sale,allow_recycle,interaction_type,interaction_modes_count,multiheight,custom_params,metadata)
overriding system value values
    (4604,4604,'mnstr_seed_rare','Rare Monster Plant Seed','floor',1,1,1.00,false,false,false,false,true,true,false,true,'monsterplant_seed',2,'','8','{}'),
    (4830,4830,'mnstr_compost','RIP Monster Plant','floor',1,1,0.00,true,true,false,false,true,true,false,false,'default',8,'','','{}'),
    (14388,14388,'mnstr_seed_nt','Plant Seed (non tradeable)','floor',1,1,1.00,false,false,false,false,true,false,false,false,'monsterplant_seed',2,'','','{}'),
    (14389,14389,'mnstr_fert_nt','Monster Plant Fertiliser (non tradeable)','floor',1,1,1.00,false,false,false,false,true,false,false,false,'default',2,'','16','{}'),
    (14390,14390,'mnstr_revival_nt','Plant Revival Potion (non tradeable)','floor',1,1,0.80,true,false,false,false,true,false,false,false,'default',2,'','16','{}')
on conflict(id) do update set sprite_id=excluded.sprite_id,name=excluded.name,public_name=excluded.public_name,
    kind=excluded.kind,width=excluded.width,length=excluded.length,stack_height=excluded.stack_height,
    allow_stack=excluded.allow_stack,allow_walk=excluded.allow_walk,allow_sit=excluded.allow_sit,allow_lay=excluded.allow_lay,
    allow_inventory_stack=excluded.allow_inventory_stack,allow_trade=excluded.allow_trade,
    allow_marketplace_sale=excluded.allow_marketplace_sale,allow_recycle=excluded.allow_recycle,
    interaction_type=excluded.interaction_type,interaction_modes_count=excluded.interaction_modes_count,
    multiheight=excluded.multiheight,custom_params=excluded.custom_params,metadata=excluded.metadata,deleted_at=null,updated_at=now();

select setval(pg_get_serial_sequence('furniture_definitions','id'),greatest((select max(id) from furniture_definitions),1));
--rollback delete from furniture_definitions where id in (4604,4830,14388,14389,14390);
