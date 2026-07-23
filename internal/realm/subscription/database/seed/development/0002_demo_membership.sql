--liquibase formatted sql

--changeset pixels:pixels-subscription-seed-development-0002-demo-membership context:development
insert into subscription_memberships (
    player_id, level, started_at, expires_at, last_payday_at,
    lifetime_active_seconds, gifts_earned, gifts_claimed, version
)
values (
    1, 2, now() - interval '31 days', now() + interval '30 days', now(),
    2678400, 1, 0, 1
)
on conflict (player_id) do update set
    level = excluded.level,
    started_at = coalesce(subscription_memberships.started_at, excluded.started_at),
    expires_at = excluded.expires_at,
    last_payday_at = coalesce(subscription_memberships.last_payday_at, excluded.last_payday_at),
    lifetime_active_seconds = greatest(subscription_memberships.lifetime_active_seconds, excluded.lifetime_active_seconds),
    gifts_earned = greatest(subscription_memberships.gifts_earned, excluded.gifts_earned),
    version = subscription_memberships.version + 1;
--rollback delete from subscription_memberships where player_id = 1;
