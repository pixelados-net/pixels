--liquibase formatted sql

--changeset pixels:pixels-player-seed-development-0003-wired-actors context:development
--validCheckSum:ANY
insert into player_badges(player_id,code,equipped,source)
select id,'ACH_WiredQA1',false,'seed'
from players where lower(username) in ('milo','juno')
on conflict(player_id,code) do update set equipped=excluded.equipped,source=excluded.source;

insert into player_respect_totals(player_id,received)
select id,0 from players where lower(username) in ('milo','juno','reid','wren')
on conflict(player_id) do nothing;
--rollback delete from player_respect_totals where player_id in (1,2,3,4) and received=0; delete from player_badges where code='ACH_WiredQA1' and player_id in (1,2);
