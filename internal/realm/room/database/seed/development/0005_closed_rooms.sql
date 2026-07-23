--liquibase formatted sql

--changeset pixels:pixels-room-seed-development-0005-closed-rooms context:development
insert into rooms (id, owner_player_id, owner_name, name, description, model_name, password_hash, door_mode, max_users, score, category_id, trade_mode, staff_picked)
overriding system value
values
    (4, 4, 'carol', 'Doorbell Studio', 'A locked room for testing owner approval.', 'model_a', null, 1, 25, 4, 3, 0, false),
    (5, 1, 'demo', 'Secret Lounge', 'A password-protected social room.', 'model_b', '$2a$10$091/2yC88viGB0wJpKJJQeYYGbl3oc0qMBNnZvbnirZoa17NQlCYm', 2, 25, 6, 1, 0, false),
    (6, 2, 'alice', 'Private Workshop', 'A password-protected building room.', 'model_c', '$2a$10$091/2yC88viGB0wJpKJJQeYYGbl3oc0qMBNnZvbnirZoa17NQlCYm', 2, 15, 3, 3, 0, false),
    (7, 3, 'bob', 'Invisible Lab', 'A room visible only to authorized entrants.', 'model_b', null, 3, 15, 2, 1, 0, false)
on conflict do nothing;

insert into room_tags (room_id, tag)
values
    (4, 'doorbell'),
    (4, 'locked'),
    (5, 'password'),
    (5, 'social'),
    (6, 'password'),
    (6, 'builds'),
    (7, 'invisible')
on conflict do nothing;

select setval(
    pg_get_serial_sequence('rooms', 'id'),
    greatest((select coalesce(max(id), 1) from rooms), 1),
    true
);
--rollback delete from room_tags where room_id in (4, 5, 6, 7);
--rollback delete from rooms where id in (4, 5, 6, 7);
--rollback select setval(pg_get_serial_sequence('rooms', 'id'), greatest((select coalesce(max(id), 1) from rooms), 1), true);
