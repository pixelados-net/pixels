--liquibase formatted sql

--changeset pixels:pixels-furniture-seed-development-0012-effect-furniture context:development
update furniture_definitions set interaction_type='effect_giver',effect_pool=array[101,102,103],updated_at=now() where id=1;
update furniture_definitions set interaction_type='effect_tile',effect_male=201,effect_female=202,allow_walk=true,updated_at=now() where id=3;

--rollback update furniture_definitions set interaction_type='default',effect_pool=null,updated_at=now() where id=1;
--rollback update furniture_definitions set interaction_type='default',effect_male=null,effect_female=null,updated_at=now() where id=3;
