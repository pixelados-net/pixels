--liquibase formatted sql

--changeset pixels:messenger-seed-test-0001 context:test
insert into messenger_friendships (player_id, friend_player_id, relation)
values (1, 2, 0), (2, 1, 0)
on conflict do nothing;

--rollback delete from messenger_friendships where (player_id,friend_player_id) in ((1,2),(2,1));
