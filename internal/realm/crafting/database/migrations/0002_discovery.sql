--liquibase formatted sql
--changeset pixels:pixels-crafting-0002-discovery
create table player_crafting_known_recipes (
    player_id bigint not null references players(id) on delete cascade,
    recipe_id bigint not null references crafting_recipes(id) on delete cascade,
    discovered_at timestamptz not null default now(), primary key(player_id,recipe_id)
);
create index player_crafting_known_player_idx on player_crafting_known_recipes(player_id,discovered_at,recipe_id);
--rollback drop table if exists player_crafting_known_recipes;
