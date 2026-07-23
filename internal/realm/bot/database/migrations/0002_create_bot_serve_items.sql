--liquibase formatted sql
--changeset pixels:pixels-bot-0002-create-bot-serve-items
create table bot_serve_items (
    id bigint generated always as identity primary key,
    keyword text not null,
    definition_id bigint not null references furniture_definitions(id),
    constraint bot_serve_items_keyword_uidx unique (keyword),
    constraint bot_serve_items_keyword_chk check (keyword = lower(btrim(keyword)) and keyword <> '')
);
--rollback drop table if exists bot_serve_items;
