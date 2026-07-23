--liquibase formatted sql

--changeset pixels:pixels-player-seed-development-0001-demo-player
--validCheckSum:ANY
insert into players (id, username)
overriding system value
values
(1, 'milo'),
(2, 'juno'),
(3, 'reid'),
(4, 'wren')
on conflict do nothing;

insert into player_profiles (player_id, look, gender, motto)
values
(1, 'hr-100.hd-180-1.ch-210-66.lg-270-82.sh-290-80', 'M', 'Welcome to Pixels - make yourself at home.'),
(2, 'hr-515-45.hd-600-1.ch-665-92.lg-700-64.sh-735-68', 'F', 'Building rooms and memories.'),
(3, 'hr-828-61.hd-180-8.ch-255-81.lg-280-64.sh-305-62', 'M', 'Always down for a good chat.'),
(4, 'hr-545-33.hd-605-1.ch-635-70.lg-720-82.sh-730-92', 'F', 'Finding the best path.')
on conflict do nothing;
--rollback delete from player_profiles where player_id in (1, 2, 3, 4);
--rollback delete from players where id in (1, 2, 3, 4);
