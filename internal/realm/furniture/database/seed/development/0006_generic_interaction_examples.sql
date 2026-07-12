--liquibase formatted sql

--changeset pixels:pixels-furniture-seed-development-0006-generic-interaction-examples context:development
insert into furniture_definitions (
    id, sprite_id, name, public_name, kind, width, length, stack_height,
    allow_stack, allow_walk, allow_sit, allow_lay, allow_inventory_stack,
    interaction_type, interaction_modes_count, multiheight, custom_params, metadata
)
overriding system value
values
    (25, 3886, 'tv_flat', 'Flatscreen TV', 'floor', 2, 1, 1.00, true, false, false, false, true, 'default', 2, '', '', '{}'),
    (26, 57, 'lamp_armas', 'Table Lamp', 'floor', 1, 1, 1.00, true, false, false, false, true, 'default', 2, '', '', '{}'),
    (27, 3676, 'wf_colorwheel', 'Color Wheel', 'floor', 1, 1, 1.00, true, false, false, false, true, 'default', 10, '', '', '{}')
on conflict (id) do update set
    sprite_id = excluded.sprite_id,
    name = excluded.name,
    public_name = excluded.public_name,
    kind = excluded.kind,
    width = excluded.width,
    length = excluded.length,
    stack_height = excluded.stack_height,
    allow_stack = excluded.allow_stack,
    allow_walk = excluded.allow_walk,
    allow_sit = excluded.allow_sit,
    allow_lay = excluded.allow_lay,
    allow_inventory_stack = excluded.allow_inventory_stack,
    interaction_type = excluded.interaction_type,
    interaction_modes_count = excluded.interaction_modes_count,
    updated_at = now();

select setval(pg_get_serial_sequence('furniture_definitions', 'id'), greatest((select max(id) from furniture_definitions), 1));
--rollback delete from furniture_definitions where id in (25, 26, 27);
