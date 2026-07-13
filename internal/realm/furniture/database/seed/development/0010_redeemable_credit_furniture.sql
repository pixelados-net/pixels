--liquibase formatted sql

--changeset pixels:pixels-furniture-seed-development-0010-redeemable-credit-furniture context:development
insert into furniture_definitions (id,sprite_id,name,public_name,kind,width,length,stack_height,allow_stack,allow_walk,allow_sit,allow_lay,allow_inventory_stack,allow_trade,allow_marketplace_sale,redeemable_credits,interaction_type,interaction_modes_count,multiheight,custom_params,metadata)
overriding system value
values (2000,2000,'credit_furni_50','Credit Furni (50)','floor',1,1,1,false,false,false,false,true,true,false,50,'default',1,'','','{}')
on conflict (id) do update set allow_trade=true,allow_marketplace_sale=false,redeemable_credits=50;

insert into furniture_items(definition_id,owner_player_id,extra_data)
select 2000,1,'0' where not exists(select 1 from furniture_items where definition_id=2000 and owner_player_id=1 and deleted_at is null);

select setval(pg_get_serial_sequence('furniture_definitions','id'),greatest((select max(id) from furniture_definitions),1));
--rollback delete from furniture_items where definition_id=2000; delete from furniture_definitions where id=2000;
