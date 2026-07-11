--liquibase formatted sql

--changeset pixels:pixels-player-seed-development-0002-demo-club context:development
update players
set club_level = 2, club_expires_at = now() + interval '30 days', updated_at = now()
where id = 1;
--rollback update players set club_level = 0, club_expires_at = null, updated_at = now() where id = 1;
