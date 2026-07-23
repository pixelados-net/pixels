--liquibase formatted sql
--changeset pixels:pixels-progression-0005-quizzes-promos
create table quizzes (
    code text primary key, kind text not null check(kind in('safety','poll')), enabled boolean not null default true
);
create table quiz_questions (
    id bigint generated always as identity primary key, quiz_code text not null references quizzes(code) on delete cascade,
    question_ref integer not null, correct_answer_id integer not null, unique(quiz_code,question_ref)
);
create table player_quiz_results (
    player_id bigint not null references players(id) on delete cascade, quiz_code text not null references quizzes(code) on delete cascade,
    passed boolean not null, failed_question_refs integer[] not null default '{}', attempted_at timestamptz not null default now(),
    passed_at timestamptz null, primary key(player_id,quiz_code)
);
create table promo_badges (
    code text primary key, badge_code text not null, starts_at timestamptz null, ends_at timestamptz null,
    max_claims bigint not null default 0 check(max_claims>=0), enabled boolean not null default true,
    constraint promo_badge_window_chk check(starts_at is null or ends_at is null or starts_at<ends_at)
);
create table promo_badge_claims (
    player_id bigint not null references players(id) on delete cascade, code text not null references promo_badges(code) on delete cascade,
    claimed_at timestamptz not null default now(), primary key(player_id,code)
);
create index promo_badge_claims_code_idx on promo_badge_claims(code,claimed_at,player_id);
--rollback drop table if exists promo_badge_claims; drop table if exists promo_badges; drop table if exists player_quiz_results; drop table if exists quiz_questions; drop table if exists quizzes;
