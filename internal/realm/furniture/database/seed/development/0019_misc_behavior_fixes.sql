--liquibase formatted sql

--changeset pixels:pixels-furniture-seed-development-0019-misc-behavior-fixes context:development
update furniture_definitions
set sprite_id = case id when 30301 then 1512 when 30302 then 1519 when 30303 then 1526 end,
    name = case id when 30301 then 'prizetrophy*1' when 30302 then 'prizetrophy*2' when 30303 then 'prizetrophy*3' end,
    public_name = case id when 30301 then 'Classic Gold Trophy' when 30302 then 'Classic Silver Trophy' when 30303 then 'Classic Bronze Trophy' end,
    interaction_type = 'trophy',
    interaction_modes_count = 1,
    stack_height = 1.00,
    updated_at = now()
where id between 30301 and 30303;

update furniture_definitions
set interaction_type = 'gate',
    interaction_modes_count = 2,
    allow_walk = false,
    custom_params = 'open_state=0',
    updated_at = now()
where id between 30038 and 30049
  and name like 'bazaar_c17_curtain%';

--rollback update furniture_definitions set sprite_id=case id when 30301 then 185 when 30302 then 186 when 30303 then 243 end,name=case id when 30301 then 'prize1' when 30302 then 'prize2' when 30303 then 'prize3' end,public_name=case id when 30301 then 'Gold Trophy' when 30302 then 'Silver Trophy' when 30303 then 'Bronze Trophy' end,interaction_type='trophy',interaction_modes_count=2,stack_height=1.00,updated_at=now() where id between 30301 and 30303; update furniture_definitions set interaction_type=case when id in (30038,30040,30041,30043) then 'gate' else 'default' end,interaction_modes_count=case when id=30038 then 1 else 2 end,allow_walk=false,custom_params='',updated_at=now() where id between 30038 and 30049;
