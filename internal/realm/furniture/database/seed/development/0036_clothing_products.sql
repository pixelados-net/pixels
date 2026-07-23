--liquibase formatted sql

--changeset pixels:pixels-furniture-seed-development-0036-clothing-products context:development
--validCheckSum:ANY
insert into furniture_definitions (id,sprite_id,name,public_name,kind,width,length,stack_height,allow_stack,allow_walk,allow_sit,allow_lay,allow_inventory_stack,allow_trade,allow_marketplace_sale,interaction_type,interaction_modes_count,multiheight,custom_params,metadata)
overriding system value values
    (6284,6284,'clothing_squid','Squid Hat','floor',1,1,0,true,false,false,false,true,true,false,'clothing',2,'','','{}')
on conflict(id) do update set sprite_id=excluded.sprite_id,name=excluded.name,public_name=excluded.public_name,
    interaction_type=excluded.interaction_type,interaction_modes_count=excluded.interaction_modes_count,deleted_at=null,updated_at=now();

insert into clothing_products(product_code,definition_id,enabled)
values('clothing_squid',6284,true)
on conflict(product_code) do update set definition_id=excluded.definition_id,enabled=true;

insert into clothing_product_sets(product_code,figure_set_id)
values('clothing_squid',3356)
on conflict do nothing;

insert into furniture_items(definition_id,owner_player_id,extra_data,metadata)
select 6284,p.id,'0','{"seed":"user_remaining_clothing"}'::jsonb
from players p
where lower(p.username)='milo'
  and not exists(select 1 from furniture_items fi where fi.owner_player_id=p.id and fi.definition_id=6284 and fi.deleted_at is null);

select setval(pg_get_serial_sequence('furniture_definitions','id'),greatest((select max(id) from furniture_definitions),1));

--rollback delete from furniture_items where metadata->>'seed'='user_remaining_clothing'; delete from clothing_product_sets where product_code='clothing_squid'; delete from clothing_products where product_code='clothing_squid'; delete from furniture_definitions where id=6284;
