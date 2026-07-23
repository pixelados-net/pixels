--liquibase formatted sql

--changeset pixels:pixels-camera-seed-development-0001-camera context:development
insert into furniture_definitions (
    id, sprite_id, name, public_name, kind, width, length, stack_height,
    allow_stack, allow_walk, allow_sit, allow_lay, allow_inventory_stack,
    allow_trade, allow_marketplace_sale, allow_recycle, interaction_type,
    interaction_modes_count, multiheight, custom_params, metadata
)
overriding system value
values (
    940001, 4536, 'external_image_wallitem_poster_small', 'Foto', 'wall', 1, 1, 0.00,
    false, false, false, false, false, true, true, false, 'external_image',
    1, '', '', '{}'
)
on conflict (id) do update set name=excluded.name, public_name=excluded.public_name,
    interaction_type=excluded.interaction_type, allow_recycle=false;

insert into camera_settings (id, publish_cooldown_seconds)
values (1, 10)
on conflict (id) do update set publish_cooldown_seconds=excluded.publish_cooldown_seconds;

select setval(pg_get_serial_sequence('furniture_definitions', 'id'), greatest((select max(id) from furniture_definitions), 1));
--rollback delete from furniture_definitions where id=940001;
