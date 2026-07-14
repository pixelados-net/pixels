--liquibase formatted sql

--changeset pixels:pixels-catalog-seed-development-0013-effect-asset-ids context:development
update catalog_items set name='effect_twinkle_permanent',grants_effect_id=4,updated_at=now() where id=1201;
update catalog_items set name='effect_torch_1day',grants_effect_id=5,updated_at=now() where id=1202;
update catalog_items set name='effect_hc_gold_spotlight_permanent',grants_effect_id=137,updated_at=now() where id=1203;

--rollback update catalog_items set name='effect_confetti_permanent',grants_effect_id=101,updated_at=now() where id=1201;
--rollback update catalog_items set name='effect_flames_1day',grants_effect_id=103,updated_at=now() where id=1202;
--rollback update catalog_items set name='effect_hc_aura_permanent',grants_effect_id=201,updated_at=now() where id=1203;
