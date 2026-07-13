--liquibase formatted sql

--changeset pixels:pixels-player-0006-add-allow-trade
alter table players add column allow_trade boolean not null default true;
--rollback alter table players drop column if exists allow_trade;
