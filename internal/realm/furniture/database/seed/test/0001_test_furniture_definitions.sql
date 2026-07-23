--liquibase formatted sql

--changeset pixels:pixels-furniture-seed-test-0001-test-furniture-definitions context:test
insert into furniture_definitions (id, sprite_id, name, public_name, kind, width, length, stack_height, allow_stack, allow_walk, allow_sit, allow_lay, allow_inventory_stack, interaction_type, interaction_modes_count, multiheight, custom_params, metadata)
overriding system value
values
    (1, 22, 'table_plasto_4leg', 'table_plasto_4leg', 'floor', 2, 2, 1.00, true, false, false, false, true, 'default', 2, '', '', '{}'),
    (2, 39, 'chair_plasto', 'chair_plasto', 'floor', 1, 1, 1.00, false, false, true, false, true, 'default', 2, '', '', '{"slots":[{"dx":0,"dy":0,"status":"sit","body_rotation":0}]}'),
    (3, 28, 'sofa_silo', 'Gray Sofa', 'floor', 2, 1, 1.10, true, false, true, false, true, 'default', 2, '', '', '{"slots":[{"dx":0,"dy":0,"status":"sit","body_rotation":0},{"dx":1,"dy":0,"status":"sit","body_rotation":0}]}'),
    (4, 45, 'bed_silo_one', 'Single Bed', 'floor', 1, 3, 1.80, false, false, false, true, true, 'bed', 1, '', '', '{"slots":[{"dx":0,"dy":0,"status":"lay","body_rotation":0}]}'),
    (5, 46, 'bed_silo_two', 'Double Bed', 'floor', 2, 3, 1.80, false, false, false, true, true, 'bed', 1, '', '', '{"slots":[{"dx":0,"dy":0,"status":"lay","body_rotation":0},{"dx":1,"dy":0,"status":"lay","body_rotation":0}]}')
on conflict do nothing;
--rollback delete from furniture_definitions where id in (1, 2, 3, 4, 5);
