--liquibase formatted sql
--changeset pixels:pixels-crafting-0001-recipes
create table crafting_altars (
    definition_id bigint primary key references furniture_definitions(id),
    enabled boolean not null default true,
    created_at timestamptz not null default now(), updated_at timestamptz not null default now(), version bigint not null default 1 check(version>0)
);
create table crafting_recipes (
    id bigint generated always as identity primary key,
    altar_definition_id bigint not null references crafting_altars(definition_id),
    name text not null, reward_definition_id bigint not null references furniture_definitions(id),
    secret boolean not null default false, limited boolean not null default false,
    remaining integer null, achievement_code text null, enabled boolean not null default true,
    created_at timestamptz not null default now(), updated_at timestamptz not null default now(), version bigint not null default 1,
    unique(altar_definition_id,name),
    constraint crafting_recipe_name_chk check(char_length(name) between 1 and 70),
    constraint crafting_recipe_stock_chk check((limited and remaining is not null and remaining>=0) or (not limited and remaining is null)),
    constraint crafting_recipe_version_chk check(version>0)
);
create table crafting_recipe_ingredients (
    recipe_id bigint not null references crafting_recipes(id) on delete cascade,
    ingredient_definition_id bigint not null references furniture_definitions(id), amount integer not null check(amount>0),
    primary key(recipe_id,ingredient_definition_id)
);
create index crafting_recipes_altar_idx on crafting_recipes(altar_definition_id,id) where enabled;
--rollback drop table if exists crafting_recipe_ingredients; drop table if exists crafting_recipes; drop table if exists crafting_altars;
