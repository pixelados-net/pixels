--liquibase formatted sql

--changeset pixels:pixels-player-seed-development-0001-demo-player
insert into players (id, username, email, look, motto)
values (
    '10000000-0000-0000-0000-000000000001',
    'demo',
    'demo@example.test',
    'hr-100.hd-180-1.ch-210-66.lg-270-82.sh-290-80',
    'Welcome to Pixels.'
)
on conflict do nothing;
--rollback delete from players where id = '10000000-0000-0000-0000-000000000001';
