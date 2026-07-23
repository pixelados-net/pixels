--liquibase formatted sql

--changeset pixels:pixels-camera-0007-drop-legacy-pending-index
drop index if exists camera_captures_pending_idx;

--rollback create index camera_captures_pending_idx on camera_captures(player_id, created_at desc) where kind='photo' and consumed_at is null;
