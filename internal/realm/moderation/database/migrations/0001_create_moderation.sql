--liquibase formatted sql

--changeset pixels:pixels-moderation-0001-create-moderation
create table cfh_topics (
    id bigint generated always as identity primary key,
    category text not null,
    name_key text not null,
    action text not null default 'queue',
    auto_reply_key text null,
    default_sanction_ladder boolean not null default false,
    order_num integer not null default 0,
    enabled boolean not null default true,
    constraint cfh_topics_action_chk check (action in ('queue','auto_reply','ignore'))
);
create table moderation_issues (
    id bigint generated always as identity primary key,
    reporter_player_id bigint not null references players(id),
    reported_player_id bigint null references players(id),
    room_id bigint null references rooms(id),
    topic_id bigint not null references cfh_topics(id),
    kind text not null default 'cfh',
    message text not null,
    state text not null default 'open',
    resolution integer null,
    picked_by_player_id bigint null references players(id),
    picked_at timestamptz null,
    closed_by_player_id bigint null references players(id),
    closed_at timestamptz null,
    created_at timestamptz not null default now(),
    constraint moderation_issues_state_chk check (state in ('open','picked','resolved','deleted'))
);
create index moderation_issues_queue_idx on moderation_issues(state,created_at);
create index moderation_issues_reporter_idx on moderation_issues(reporter_player_id,state,created_at desc);
create index moderation_issues_reported_idx on moderation_issues(reported_player_id,created_at desc);
create table issue_chatlog (
    id bigint generated always as identity primary key,
    issue_id bigint not null references moderation_issues(id) on delete cascade,
    player_id bigint null references players(id),
    pattern_id text not null default '',
    message text not null,
    created_at timestamptz not null default now()
);
create index issue_chatlog_issue_idx on issue_chatlog(issue_id,id);
create table moderation_presets (
    id bigint generated always as identity primary key,
    category text not null,
    message_key text not null,
    enabled boolean not null default true,
    order_num integer not null default 0
);
create table moderator_preferences (
    player_id bigint primary key references players(id),
    window_x integer not null default 0,
    window_y integer not null default 0,
    window_width integer not null default 640,
    window_height integer not null default 480,
    updated_at timestamptz not null default now()
);
create table guide_feedback (
    id bigint generated always as identity primary key,
    guide_player_id bigint not null references players(id),
    requester_player_id bigint not null references players(id),
    recommended boolean not null,
    created_at timestamptz not null default now()
);
create table guardian_tickets (
    id bigint generated always as identity primary key,
    reporter_player_id bigint not null references players(id),
    reported_player_id bigint not null references players(id),
    state text not null default 'offered',
    result integer null,
    created_at timestamptz not null default now(),
    closes_at timestamptz not null
);
create table guardian_votes (
    ticket_id bigint not null references guardian_tickets(id) on delete cascade,
    guardian_player_id bigint not null references players(id),
    vote integer not null,
    created_at timestamptz not null default now(),
    primary key(ticket_id,guardian_player_id),
    constraint guardian_votes_vote_chk check (vote between 0 and 2)
);
alter table punishments add constraint punishments_cfh_topic_fk foreign key(cfh_topic_id) references cfh_topics(id);
alter table punishments add constraint punishments_issue_fk foreign key(issue_id) references moderation_issues(id);
insert into cfh_topics(category,name_key,action,auto_reply_key,default_sanction_ladder,order_num) values
    ('harassment','moderation.topic.harassment','queue',null,true,10),
    ('scam','moderation.topic.scam','queue',null,true,20),
    ('help','moderation.topic.help','auto_reply','moderation.reply.help',false,30),
    ('other','moderation.topic.other','queue',null,false,40);
insert into moderation_presets(category,message_key,order_num) values
    ('general','moderation.preset.warning',10),
    ('general','moderation.preset.resolved',20);
--rollback alter table punishments drop constraint if exists punishments_issue_fk; alter table punishments drop constraint if exists punishments_cfh_topic_fk; drop table if exists guardian_votes; drop table if exists guardian_tickets; drop table if exists guide_feedback; drop table if exists moderator_preferences; drop table if exists moderation_presets; drop table if exists issue_chatlog; drop table if exists moderation_issues; drop table if exists cfh_topics;
