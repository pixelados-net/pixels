--liquibase formatted sql

--changeset pixels:permission-groups-badge-url
ALTER TABLE permission_groups
    ADD COLUMN badge_url TEXT NOT NULL DEFAULT '';

ALTER TABLE permission_groups
    ADD CONSTRAINT permission_groups_badge_url_length
        CHECK (char_length(badge_url) <= 2048);

--rollback ALTER TABLE permission_groups DROP CONSTRAINT permission_groups_badge_url_length;
--rollback ALTER TABLE permission_groups DROP COLUMN badge_url;
