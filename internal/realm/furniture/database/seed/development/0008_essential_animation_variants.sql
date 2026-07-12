--liquibase formatted sql

--changeset pixels:pixels-furniture-seed-development-0008-essential-animation-variants context:development
insert into furniture_definitions (
    id, sprite_id, name, public_name, kind, width, length, stack_height,
    allow_stack, allow_walk, allow_sit, allow_lay, allow_inventory_stack,
    interaction_type, interaction_modes_count, multiheight, custom_params, metadata
)
overriding system value
values
    (31, 2694, 'fortune_random_state_test', 'Magic Crystal Ball (Random State)', 'floor', 1, 1, 1.00, true, false, false, false, true, 'random_state', 9, '', 'states=9,delay=1200', '{"seed":"arcturus:fortune","test_variant":true}'),
    (32, 3693, 'wf_arrowplate_color_test', 'Arrow Color Plate', 'floor', 1, 1, 0.10, true, true, false, false, true, 'colorplate', 4, '', '', '{"seed":"arcturus:wf_arrowplate","test_variant":true}'),
    (33, 3701, 'wf_floor_switch_remote_test', 'Remote Floor Switch', 'floor', 1, 1, 1.00, true, false, false, false, true, 'switch_remote_control', 2, '', '', '{"seed":"arcturus:wf_floor_switch1","test_variant":true}'),
    (34, 201, 'fridge_no_sides_test', 'Pura Refrigerator (All Sides)', 'floor', 1, 1, 1.00, true, false, false, false, true, 'vendingmachine_no_sides', 2, '', '3,5,6,2,4', '{"seed":"arcturus:fridge","test_variant":true}'),
    (35, 177, 'sink_handitem_test', 'Sink Hand Item Dispenser', 'floor', 1, 1, 1.00, true, false, false, false, true, 'handitem', 2, '', '100', '{"seed":"arcturus:sink","test_variant":true}')
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
    custom_params = excluded.custom_params,
    metadata = excluded.metadata,
    updated_at = now();

select setval(pg_get_serial_sequence('furniture_definitions', 'id'), greatest((select max(id) from furniture_definitions), 1));
--rollback delete from furniture_definitions where id in (31, 32, 33, 34, 35);
