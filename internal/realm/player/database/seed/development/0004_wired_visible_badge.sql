--liquibase formatted sql

--changeset pixels:pixels-player-seed-development-0004-wired-visible-badge context:development
delete from player_badges
where upper(code)='ACH_WIREDQA1'
  and player_id in (select id from players where lower(username) in ('demo','alice'));

insert into player_badges(player_id,code,equipped,slot,source)
select id,'ADM',lower(username)='demo',case when lower(username)='demo' then 1 else null end,'seed'
from players
where lower(username) in ('demo','alice')
on conflict(player_id,code) do update set
    equipped=excluded.equipped,
    slot=excluded.slot,
    source=excluded.source;

--rollback delete from player_badges where code='ADM' and source='seed' and player_id in (select id from players where lower(username) in ('demo','alice'));
