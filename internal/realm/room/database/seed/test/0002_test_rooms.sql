--liquibase formatted sql

--changeset pixels:pixels-room-seed-test-0002-test-rooms context:test
insert into room_categories (id, caption, caption_key, visible, order_num)
overriding system value
values
    (1, 'Test Social', 'test_social', true, 1)
on conflict do nothing;

insert into rooms (id, owner_player_id, owner_name, name, description, model_name, max_users, score, category_id, trade_mode)
overriding system value
values
    (1, 1, 'test_demo', 'Test Room', 'A test navigator fixture.', 'model_test', 25, 1, 1, 0)
on conflict do nothing;

insert into room_tags (room_id, tag)
values
    (1, 'test')
on conflict do nothing;
--rollback delete from room_tags where room_id = 1;
--rollback delete from rooms where id = 1;
--rollback delete from room_categories where id = 1;
