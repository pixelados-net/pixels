--liquibase formatted sql

--changeset pixels:pixels-room-seed-development-0002-default-rooms context:development
--validCheckSum:ANY
insert into room_categories (id, caption, caption_key, visible, order_num)
overriding system value
values
    (1, 'Social', 'social', true, 1),
    (2, 'Games', 'games', true, 2),
    (3, 'Builds', 'builds', true, 3)
on conflict do nothing;

insert into rooms (id, owner_player_id, owner_name, name, description, model_name, max_users, score, category_id, trade_mode, staff_picked)
overriding system value
values
    (1, 1, 'milo', 'Pixels Lobby', 'The heart of the hotel - welcome new arrivals and catch up with friends.', 'model_a', 25, 15, 1, 0, true),
    (2, 2, 'juno', 'Builder Corner', 'Furni swaps, room tours, and works in progress.', 'model_b', 25, 8, 3, 0, false),
    (3, 3, 'reid', 'The Chat Lounge', 'Good music and even better conversation.', 'model_c', 15, 5, 1, 0, false)
on conflict do nothing;

insert into room_tags (room_id, tag)
values
    (1, 'lobby'),
    (1, 'social'),
    (2, 'builds'),
    (3, 'lounge')
on conflict do nothing;
--rollback delete from room_tags where room_id in (1, 2, 3);
--rollback delete from rooms where id in (1, 2, 3);
--rollback delete from room_categories where id in (1, 2, 3);
