--liquibase formatted sql

--changeset pixels:pixels-furniture-seed-development-0014-room-decorators context:development
insert into furniture_definitions (id,sprite_id,name,public_name,kind,width,length,stack_height,allow_stack,allow_walk,allow_sit,allow_lay,allow_inventory_stack,interaction_type,interaction_modes_count,multiheight,custom_params,metadata)
overriding system value values
    (40,3983,'note_tag','Stickie Pole','floor',1,1,1.00,true,false,false,false,true,'sticky_pole',2,'','','{}'),
    (41,4170,'boutique_mannequin1','Mannequin','floor',1,1,0.00,true,false,false,false,true,'mannequin',1,'','','{}'),
    (42,4697,'roombg_color','Background Toner','floor',1,1,0.45,true,true,false,false,true,'background_toner',2,'','','{}')
on conflict(id) do update set sprite_id=excluded.sprite_id,name=excluded.name,public_name=excluded.public_name,kind=excluded.kind,interaction_type=excluded.interaction_type,interaction_modes_count=excluded.interaction_modes_count;
select setval(pg_get_serial_sequence('furniture_definitions','id'),greatest((select max(id) from furniture_definitions),1));
--rollback delete from furniture_definitions where id in (40,41,42);
