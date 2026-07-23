--liquibase formatted sql

--changeset pixels:messenger-seed-development-0001 context:development
insert into messenger_friendships (player_id, friend_player_id, relation)
values
    (1, 2, 1), (2, 1, 2),
    (1, 3, 0), (3, 1, 0),
    (2, 4, 3), (4, 2, 0)
on conflict (player_id, friend_player_id) do update
set relation = excluded.relation;

insert into messenger_friend_requests (from_player_id, to_player_id)
values (4, 1), (3, 2)
on conflict do nothing;

--rollback delete from messenger_friend_requests where (from_player_id,to_player_id) in ((4,1),(3,2)); delete from messenger_friendships where (player_id,friend_player_id) in ((1,2),(2,1),(1,3),(3,1),(2,4),(4,2));
