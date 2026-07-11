--liquibase formatted sql

--changeset pixels:pixels-player-seed-test-0002-test-club context:test
update players
set club_level = 2, club_expires_at = now() + interval '1 day', updated_at = now()
where id = 1;
--rollback update players set club_level = 0, club_expires_at = null, updated_at = now() where id = 1;
