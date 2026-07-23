--liquibase formatted sql

--changeset pixels:pixels-furniture-seed-development-0007-essential-interactions context:development
insert into furniture_definitions (
    id, sprite_id, name, public_name, kind, width, length, stack_height,
    allow_stack, allow_walk, allow_sit, allow_lay, allow_inventory_stack,
    interaction_type, interaction_modes_count, multiheight, custom_params, metadata
)
overriding system value
values
    (28, 3688, 'wf_pressureplate', 'Pressure Plate', 'floor', 1, 1, 0.20, true, true, false, false, true, 'pressureplate', 2, '', '', '{}'),
    (29, 3701, 'wf_floor_switch1', 'Floor Switch', 'floor', 1, 1, 1.00, true, false, false, false, true, 'switch', 2, '', '', '{}'),
    (30, 4192, 'eleblock2', 'Pink Pura Block', 'floor', 1, 1, 0.01, true, true, false, false, true, 'multiheight', 5, '0.01;0.60;0.99;1.80;2.00', '', '{}')
on conflict (id) do update set
    sprite_id = excluded.sprite_id,
    name = excluded.name,
    public_name = excluded.public_name,
    width = excluded.width,
    length = excluded.length,
    stack_height = excluded.stack_height,
    allow_stack = excluded.allow_stack,
    allow_walk = excluded.allow_walk,
    interaction_type = excluded.interaction_type,
    interaction_modes_count = excluded.interaction_modes_count,
    multiheight = excluded.multiheight,
    updated_at = now();

update furniture_definitions
set custom_params = '2,5,7', allow_walk = true, stack_height = 0, updated_at = now()
where id = 24 and interaction_type = 'handitem_tile';

select setval(pg_get_serial_sequence('furniture_definitions', 'id'), greatest((select max(id) from furniture_definitions), 1));
--rollback update furniture_definitions set custom_params = '', allow_walk = false, stack_height = 1, updated_at = now() where id = 24;
--rollback delete from furniture_definitions where id in (28, 29, 30);
