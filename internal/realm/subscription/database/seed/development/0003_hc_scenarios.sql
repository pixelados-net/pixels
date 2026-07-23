--liquibase formatted sql

--changeset pixels:pixels-subscription-seed-development-0003-hc-scenarios context:development
insert into subscription_memberships (
    player_id,level,started_at,streak_started_at,expires_at,last_payday_at,last_accrued_at,
    lifetime_active_seconds,lifetime_vip_seconds,gifts_earned,gifts_claimed,version
)
values
    (1,1,now()-interval '70 days',now()-interval '70 days',now()+interval '30 days',now()-interval '32 days',now()-interval '2 days',2678400,0,1,0,1),
    (2,2,now()-interval '100 days',now()-interval '46 days',now()+interval '45 days',now()-interval '10 days',now()-interval '1 hour',3888000,3888000,1,0,1),
    (3,1,now()-interval '70 days',now()-interval '70 days',now()+interval '30 days',now()-interval '63 days',now()-interval '5 days',5011200,0,1,0,1),
    (4,0,now()-interval '90 days',null,null,now()-interval '20 days',now()-interval '20 days',2678400,0,1,1,1)
on conflict (player_id) do update set
    level=excluded.level,started_at=excluded.started_at,streak_started_at=excluded.streak_started_at,
    expires_at=excluded.expires_at,last_payday_at=excluded.last_payday_at,last_accrued_at=excluded.last_accrued_at,
    lifetime_active_seconds=excluded.lifetime_active_seconds,lifetime_vip_seconds=excluded.lifetime_vip_seconds,gifts_earned=excluded.gifts_earned,
    gifts_claimed=excluded.gifts_claimed,version=subscription_memberships.version+1;

update players p set
    club_level=m.level,
    club_expires_at=m.expires_at,
    updated_at=now()
from subscription_memberships m
where p.id=m.player_id and p.id in (1,2,3,4);

with granted as (
    insert into furniture_items (definition_id,owner_player_id,extra_data,metadata)
    values (3,1,'0','{"seed":"hc_payday_purchase"}'::jsonb)
    returning id
), purchased as (
    insert into catalog_purchase_log (player_id,catalog_item_id,quantity,cost_credits,cost_points,points_type,purchased_at)
    values (1,3,1,12,0,-1,now()-interval '20 days')
    returning id
)
insert into catalog_purchase_items (purchase_id,furniture_item_id)
select purchased.id,granted.id from purchased cross join granted;

with granted as (
    insert into furniture_items (definition_id,owner_player_id,extra_data,metadata)
    select 3,3,'0','{"seed":"hc_catchup_purchase"}'::jsonb from generate_series(1,5)
    returning id
), purchased as (
    insert into catalog_purchase_log (player_id,catalog_item_id,quantity,cost_credits,cost_points,points_type,purchased_at)
    values (3,3,5,60,0,-1,now()-interval '50 days')
    returning id
)
insert into catalog_purchase_items (purchase_id,furniture_item_id)
select purchased.id,granted.id from purchased cross join granted;

with granted as (
    insert into furniture_items (definition_id,owner_player_id,extra_data,metadata)
    select 2,3,'0','{"seed":"hc_catchup_purchase"}'::jsonb from generate_series(1,4)
    returning id
), purchased as (
    insert into catalog_purchase_log (player_id,catalog_item_id,quantity,cost_credits,cost_points,points_type,purchased_at)
    values (3,1,4,8,0,-1,now()-interval '10 days')
    returning id
)
insert into catalog_purchase_items (purchase_id,furniture_item_id)
select purchased.id,granted.id from purchased cross join granted;

update subscription_targeted_offers set
    image_url='targetedoffers/ufo_habbo20_mach1.png',
    icon_url='targetedoffers/tto_blkfri_20_small.png',
    expires_at=now()+interval '30 days',
    enabled=true
where title_key='offer.targeted.dev.title';
--rollback delete from subscription_targeted_offer_progress where player_id in (1,2,3,4); delete from catalog_purchase_items where purchase_id in (select id from catalog_purchase_log where player_id in (1,3) and purchased_at>now()-interval '70 days'); delete from catalog_purchase_log where player_id in (1,3) and purchased_at>now()-interval '70 days'; delete from furniture_items where metadata->>'seed' in ('hc_payday_purchase','hc_catchup_purchase'); delete from subscription_memberships where player_id in (1,2,3,4); update players set club_level=0,club_expires_at=null,updated_at=now() where id in (1,2,3,4);
