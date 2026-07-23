--liquibase formatted sql

--changeset pixels:pixels-player-seed-test-0001-test-player
insert into players (id, username)
overriding system value
values
(1, 'test_demo'),
(2, 'test_player'),
(3, 'test_alice'),
(4, 'test_bob')
on conflict do nothing;

insert into player_profiles (player_id, look, gender, motto)
values
(1, 'hr-100.hd-180-1.ch-210-66.lg-270-82.sh-290-80', 'M', 'Test demo fixture.'),
(2, 'hr-100.hd-180-1.ch-210-66.lg-270-82.sh-290-80', 'M', 'Test fixture.'),
(3, 'hr-515-45.hd-600-1.ch-665-92.lg-700-64.sh-735-68', 'F', 'Test friend fixture.'),
(4, 'hr-828-61.hd-180-8.ch-255-81.lg-280-64.sh-305-62', 'M', 'Test room fixture.')
on conflict do nothing;
--rollback delete from player_profiles where player_id in (1, 2, 3, 4);
--rollback delete from players where id in (1, 2, 3, 4);
