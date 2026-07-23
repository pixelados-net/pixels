--liquibase formatted sql

--changeset pixels:pixels-sanction-0002-create-sanction-ladder
create table sanction_ladder (
    level integer primary key,
    kind text not null,
    duration_hours integer not null,
    probation_days integer not null,
    constraint sanction_ladder_level_chk check (level > 0),
    constraint sanction_ladder_kind_chk check (kind in ('warn','mute','ban')),
    constraint sanction_ladder_duration_chk check (duration_hours >= 0),
    constraint sanction_ladder_probation_chk check (probation_days >= 0)
);
insert into sanction_ladder(level,kind,duration_hours,probation_days) values
    (1,'warn',0,7),
    (2,'mute',1,14),
    (3,'ban',24,30),
    (4,'ban',168,60),
    (5,'ban',0,365);
--rollback drop table if exists sanction_ladder;
