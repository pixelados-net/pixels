--liquibase formatted sql

--changeset pixels:pixels-sanction-0001-create-punishments
create table punishments (
    id bigint generated always as identity primary key,
    receiver_player_id bigint not null references players(id),
    issuer_player_id bigint null references players(id),
    issuer_kind text not null default 'player',
    kind text not null,
    reason text not null,
    cfh_topic_id bigint null,
    issue_id bigint null,
    source text not null,
    issued_at timestamptz not null default now(),
    expires_at timestamptz null,
    revoked_at timestamptz null,
    revoked_by_player_id bigint null references players(id),
    constraint punishments_kind_chk check (kind in ('ban','mute','warn','trade_lock','kick')),
    constraint punishments_issuer_chk check (issuer_kind in ('player','system') and (issuer_kind = 'system' or issuer_player_id is not null)),
    constraint punishments_instant_expiry_chk check (kind not in ('warn','kick') or expires_at is null)
);
create index punishments_active_idx on punishments(receiver_player_id,kind) where revoked_at is null;
create index punishments_history_idx on punishments(receiver_player_id,issued_at desc);
create table pending_alerts (
    id bigint generated always as identity primary key,
    player_id bigint not null references players(id),
    punishment_id bigint null references punishments(id),
    message text not null,
    created_at timestamptz not null default now(),
    delivered_at timestamptz null
);
create index pending_alerts_delivery_idx on pending_alerts(player_id,id) where delivered_at is null;
insert into punishments(receiver_player_id,issuer_kind,kind,reason,source)
select id,'system','trade_lock','Migrated legacy administrative trade lock','system'
from players
where allow_trade=false;
--rollback drop table if exists pending_alerts; drop table if exists punishments;
