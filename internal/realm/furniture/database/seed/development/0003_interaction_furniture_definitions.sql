--liquibase formatted sql

--changeset pixels:pixels-furniture-seed-development-0003-interaction-definitions context:development
insert into furniture_definitions (
    id, sprite_id, name, public_name, kind, width, length, stack_height,
    allow_stack, allow_walk, allow_sit, allow_lay, allow_inventory_stack,
    interaction_type, interaction_modes_count, multiheight, custom_params, metadata
)
overriding system value
values
    (6, 244, 'divider_poly3', 'Aquamarine Gate', 'floor', 1, 1, 0.00, true, false, false, false, true, 'gate', 2, '', '', '{}'),
    (7, 1649, 'queue_tile1*3', 'Aqua Habbo Roller', 'floor', 1, 1, 0.50, true, true, false, false, true, 'roller', 1, '', '', '{}'),
    (8, 202, 'door', 'Telephone Box', 'floor', 1, 1, 0.00, true, false, false, false, true, 'teleport', 1, '', '', '{}'),
    (9, 4027, 'roomdimmer', 'Mood Light', 'wall', 1, 1, 1.00, true, false, false, false, true, 'dimmer', 1, '', '', '{}'),
    (10, 239, 'edice', 'Holodice', 'floor', 1, 1, 1.00, true, false, false, false, true, 'dice', 6, '', '', '{}'),
    (11, 127, 'bar_polyfon', 'Mini-bar', 'floor', 1, 1, 1.00, true, false, false, false, true, 'vendingmachine', 1, '', '6,5,2,1', '{}'),
    (12, 4744, 'cannon', 'Cannon', 'floor', 1, 2, 1.00, true, false, false, false, true, 'cannon', 2, '', '', '{}'),
    (13, 2597, 'one_way_door*1', 'Aquamarine One Way Gate', 'floor', 1, 1, 0.00, false, false, false, false, true, 'onewaygate', 2, '', '', '{}'),
    (14, 5315, 'info_terminal', 'Information Terminal', 'floor', 1, 1, 1.00, true, false, false, false, false, 'information_terminal', 3, '', '', '{}'),
    (15, 10840, 'mutearea_sign2', 'Mute Area Sign', 'floor', 1, 1, 0.10, true, true, false, false, false, 'mutearea', 2, '', '', '{}'),
    (16, 5520, 'buildarea_sign', 'Build Area Sign', 'floor', 1, 1, 0.00, true, false, false, false, true, 'buildarea', 2, '', '', '{}'),
    (17, 4010, 'habbowheel', 'The Wheel of Fortune', 'wall', 1, 1, 1.00, true, false, false, false, true, 'colorwheel', 10, '', '', '{}'),
    (18, 3632, 'bb_pyramid', 'Pyramid', 'floor', 1, 1, 0.00, true, false, false, false, true, 'pyramid', 1, '', '', '{}'),
    (19, 4536, 'external_image_wallitem', 'Valentines Card', 'wall', 1, 1, 0.00, true, false, false, false, true, 'external_image', 2, '', '', '{}'),
    (20, 1, 'post.it', 'Post-it Note', 'wall', 1, 1, 1.00, true, false, false, false, true, 'postit', 1, '', '', '{}'),
    (21, 3002, 'floor', 'Floor Paint', 'wall', 1, 1, 1.00, true, false, false, false, false, 'roomeffect', 1, '', '', '{}'),
    (22, 3001, 'wallpaper', 'Wallpaper', 'wall', 1, 1, 1.00, true, false, false, false, false, 'roomeffect', 1, '', '', '{}'),
    (23, 4055, 'landscape', 'Landscape', 'wall', 1, 1, 1.00, true, false, false, false, false, 'roomeffect', 1, '', '', '{}'),
    (24, 5406, 'handitem_tester', 'Hand Item Tester', 'floor', 1, 1, 1.00, true, false, false, false, true, 'handitem_tile', 2, '', '', '{}')
on conflict do nothing;

select setval(pg_get_serial_sequence('furniture_definitions', 'id'), greatest((select max(id) from furniture_definitions), 1));
--rollback delete from furniture_definitions where id between 6 and 24;
