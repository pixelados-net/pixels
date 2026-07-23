--liquibase formatted sql

--changeset pixels:room-0011-promotions
CREATE TABLE room_promotions (
    id BIGSERIAL PRIMARY KEY,
    room_id BIGINT NOT NULL UNIQUE REFERENCES rooms(id),
    category_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    starts_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ends_at TIMESTAMPTZ NOT NULL,
    created_by BIGINT NOT NULL REFERENCES players(id),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version BIGINT NOT NULL DEFAULT 1 CHECK (version > 0)
);
CREATE INDEX room_promotions_active_idx ON room_promotions (ends_at, room_id);

--rollback DROP TABLE room_promotions;
