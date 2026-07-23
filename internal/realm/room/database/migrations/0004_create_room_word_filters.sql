--liquibase formatted sql

--changeset pixels:pixels-room-0004-create-room-word-filters
create table room_word_filters (
    room_id bigint not null references rooms(id) on delete cascade,
    word text not null,
    created_at timestamptz not null default now(),
    primary key (room_id, word),
    constraint room_word_filters_word_length_chk check (char_length(word) between 1 and 32)
);

--rollback drop table if exists room_word_filters;
