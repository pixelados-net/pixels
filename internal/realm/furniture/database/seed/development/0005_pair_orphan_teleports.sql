--liquibase formatted sql

--changeset pixels:pixels-furniture-seed-development-0005-pair-orphan-teleports context:development
with unpaired as (
    select
        item.id,
        item.owner_player_id,
        item.definition_id,
        row_number() over (partition by item.owner_player_id, item.definition_id order by item.id) as position
    from furniture_items item
    join furniture_definitions definition on definition.id = item.definition_id
    where definition.interaction_type in ('teleport', 'teleport_tile')
      and not exists (
          select 1
          from furniture_item_teleport_pairs paired
          where paired.item_one_id = item.id or paired.item_two_id = item.id
      )
), pairs as (
    select first.id as item_one_id, second.id as item_two_id
    from unpaired first
    join unpaired second
      on second.owner_player_id = first.owner_player_id
     and second.definition_id = first.definition_id
     and second.position = first.position + 1
    where first.position % 2 = 1
)
insert into furniture_item_teleport_pairs (item_one_id, item_two_id)
select least(item_one_id, item_two_id), greatest(item_one_id, item_two_id)
from pairs
on conflict do nothing;
--rollback select 1;
