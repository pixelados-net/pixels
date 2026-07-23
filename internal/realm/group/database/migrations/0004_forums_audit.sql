--liquibase formatted sql

--changeset pixels:pixels-group-0004-forums-audit
create table social_group_forum_threads (
    id bigint generated always as identity primary key,
    group_id bigint not null references social_groups(id),
    author_player_id bigint not null references players(id),
    author_name text not null,
    subject text not null,
    state smallint not null default 0,
    pinned boolean not null default false,
    locked boolean not null default false,
    post_count integer not null default 0,
    last_post_id bigint not null default 0,
    last_author_player_id bigint not null,
    last_author_name text not null,
    last_posted_at timestamptz not null default now(),
    moderator_player_id bigint null references players(id),
    moderator_name text not null default '',
    moderation_reason text not null default '',
    moderated_at timestamptz null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    version bigint not null default 1,
    constraint social_group_forum_threads_subject_chk check(char_length(subject) between 1 and 120),
    constraint social_group_forum_threads_state_chk check(state in (0,1,10,20)),
    constraint social_group_forum_threads_count_chk check(post_count >= 0),
    constraint social_group_forum_threads_version_chk check(version > 0)
);

create table social_group_forum_posts (
    id bigint generated always as identity primary key,
    group_id bigint not null references social_groups(id),
    thread_id bigint not null references social_group_forum_threads(id),
    ordinal integer not null,
    author_player_id bigint not null references players(id),
    author_name text not null,
    author_figure text not null,
    body text not null,
    state smallint not null default 0,
    moderator_player_id bigint null references players(id),
    moderator_name text not null default '',
    moderation_reason text not null default '',
    moderated_at timestamptz null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    version bigint not null default 1,
    unique(thread_id,ordinal),
    constraint social_group_forum_posts_body_chk check(char_length(body) between 1 and 4000),
    constraint social_group_forum_posts_state_chk check(state in (0,1,10,20)),
    constraint social_group_forum_posts_ordinal_chk check(ordinal >= 0),
    constraint social_group_forum_posts_version_chk check(version > 0)
);

alter table social_group_forum_threads
    add constraint social_group_forum_threads_last_post_fk foreign key(last_post_id) references social_group_forum_posts(id) deferrable initially deferred;

create table social_group_forum_read_markers (
    group_id bigint not null references social_groups(id) on delete cascade,
    player_id bigint not null references players(id) on delete cascade,
    last_message_id bigint not null default 0,
    flag integer not null default 0,
    updated_at timestamptz not null default now(),
    primary key(group_id,player_id),
    constraint social_group_forum_read_markers_message_chk check(last_message_id >= 0)
);

create table social_group_forum_views (
    group_id bigint not null references social_groups(id) on delete cascade,
    player_id bigint not null references players(id) on delete cascade,
    viewed_on date not null default current_date,
    view_count integer not null default 1,
    primary key(group_id,player_id,viewed_on),
    constraint social_group_forum_views_count_chk check(view_count between 1 and 100)
);

create table social_group_audit (
    id bigint generated always as identity primary key,
    group_id bigint not null references social_groups(id),
    actor_player_id bigint null references players(id),
    action text not null,
    target_player_id bigint null references players(id),
    reason text not null default '',
    version bigint null,
    metadata jsonb not null default '{}'::jsonb,
    created_at timestamptz not null default now()
);

create index social_group_forum_threads_list_idx on social_group_forum_threads(group_id,pinned desc,last_posted_at desc,id desc);
create index social_group_forum_posts_page_idx on social_group_forum_posts(thread_id,ordinal);
create index social_group_forum_posts_author_idx on social_group_forum_posts(author_player_id,group_id);
create index social_group_forum_markers_player_idx on social_group_forum_read_markers(player_id,group_id);
create index social_group_forum_views_popular_idx on social_group_forum_views(viewed_on,group_id);
create index social_group_audit_group_idx on social_group_audit(group_id,created_at desc);

--rollback drop table if exists social_group_audit; drop table if exists social_group_forum_views; drop table if exists social_group_forum_read_markers; alter table social_group_forum_threads drop constraint if exists social_group_forum_threads_last_post_fk; drop table if exists social_group_forum_posts; drop table if exists social_group_forum_threads;
