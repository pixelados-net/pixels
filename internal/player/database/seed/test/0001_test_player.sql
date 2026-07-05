--liquibase formatted sql

--changeset pixels:pixels-player-seed-test-0001-test-player
insert into players (id, username, email, look, motto)
values (
    '20000000-0000-0000-0000-000000000001',
    'test_player',
    'test-player@example.test',
    'hr-100.hd-180-1.ch-210-66.lg-270-82.sh-290-80',
    'Test fixture.'
)
on conflict do nothing;
--rollback delete from players where id = '20000000-0000-0000-0000-000000000001';
