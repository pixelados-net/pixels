--liquibase formatted sql
--changeset pixels:pixels-crafting-seed-development-0001-lab context:development
--validCheckSum:ANY
insert into crafting_altars(definition_id,enabled) values(8388,true) on conflict(definition_id) do update set enabled=true,updated_at=now();
insert into crafting_recipes(id,altar_definition_id,name,reward_definition_id,secret,limited,remaining,achievement_code,enabled) overriding system value values
 (920001,8388,'single_bed_recipe',4,false,false,null,'Atcg',true),
 (920002,8388,'double_bed_recipe',5,false,true,1,'Atcg',true),
 (920003,8388,'gray_sofa_recipe',3,true,false,null,'AtcgSecret',true)
on conflict(id) do update set altar_definition_id=excluded.altar_definition_id,name=excluded.name,reward_definition_id=excluded.reward_definition_id,secret=excluded.secret,limited=excluded.limited,remaining=excluded.remaining,achievement_code=excluded.achievement_code,enabled=true,updated_at=now();
insert into crafting_recipe_ingredients(recipe_id,ingredient_definition_id,amount) values
 (920001,1,2),(920001,2,1),(920002,2,2),(920002,3,1),(920003,1,1),(920003,3,2)
on conflict(recipe_id,ingredient_definition_id) do update set amount=excluded.amount;
insert into crafting_recycler_prizes(tier,reward_definition_id) values(1,1),(2,4),(3,5) on conflict do nothing;
select setval(pg_get_serial_sequence('crafting_recipes','id'),greatest((select max(id) from crafting_recipes),1));
--rollback delete from crafting_recycler_prizes where (tier,reward_definition_id) in((1,1),(2,4),(3,5)); delete from crafting_recipes where id between 920001 and 920003; delete from crafting_altars where definition_id=8388;
