--liquibase formatted sql

--changeset pixels:progression-polls-0008
create table polls (
    id integer generated always as identity primary key,
    title text not null check (char_length(title) between 1 and 120),
    headline text not null default '',
    summary text not null default '',
    start_message text not null default '',
    thanks_message text not null default '',
    room_id bigint null references rooms(id) on delete set null,
    reward_badge text null,
    enabled boolean not null default true,
    version bigint not null default 1 check (version > 0),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create unique index polls_room_enabled_uidx on polls(room_id) where enabled and room_id is not null;

create table poll_questions (
    id integer generated always as identity primary key,
    poll_id integer not null references polls(id) on delete cascade,
    sort_order integer not null,
    kind integer not null check (kind between 0 and 2),
    text_ref text not null,
    category integer not null default 0,
    answer_type integer not null default 0,
    min_selections integer not null default 0,
    options jsonb not null default '[]'::jsonb,
    unique(poll_id, sort_order)
);

create table poll_answers (
    poll_id integer not null references polls(id) on delete cascade,
    question_id integer not null references poll_questions(id) on delete cascade,
    player_id bigint not null references players(id) on delete cascade,
    answer jsonb not null,
    rejected boolean not null default false,
    created_at timestamptz not null default now(),
    primary key(poll_id, question_id, player_id)
);

create index poll_answers_player_idx on poll_answers(player_id, poll_id);

--rollback drop table poll_answers;
--rollback drop table poll_questions;
--rollback drop table polls;
