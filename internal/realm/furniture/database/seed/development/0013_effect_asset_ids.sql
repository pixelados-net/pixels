--liquibase formatted sql

--changeset pixels:pixels-furniture-seed-development-0013-effect-asset-ids context:development
update furniture_definitions set effect_pool=array[4,9,113],updated_at=now() where id=1 and interaction_type='effect_giver';
update furniture_definitions set effect_male=90,effect_female=91,updated_at=now() where id=3 and interaction_type='effect_tile';

--rollback update furniture_definitions set effect_pool=array[101,102,103],updated_at=now() where id=1 and interaction_type='effect_giver';
--rollback update furniture_definitions set effect_male=201,effect_female=202,updated_at=now() where id=3 and interaction_type='effect_tile';
