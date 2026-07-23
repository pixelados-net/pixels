--liquibase formatted sql

--changeset pixels:pixels-subscription-0006-add-accrual-boundaries
alter table subscription_memberships add column streak_started_at timestamptz null;
alter table subscription_memberships add column last_accrued_at timestamptz null;
alter table subscription_memberships add column lifetime_vip_seconds bigint not null default 0;
update subscription_memberships set
    streak_started_at = coalesce(started_at, now()),
    last_accrued_at = least(now(), coalesce(expires_at, now()))
where started_at is not null;
alter table subscription_payday_log add constraint subscription_payday_log_player_period_uq unique (player_id, occurred_at);
--rollback alter table subscription_payday_log drop constraint if exists subscription_payday_log_player_period_uq; alter table subscription_memberships drop column if exists lifetime_vip_seconds; alter table subscription_memberships drop column if exists last_accrued_at; alter table subscription_memberships drop column if exists streak_started_at;
