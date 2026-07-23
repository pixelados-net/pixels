--liquibase formatted sql

--changeset pixels:pixels-subscription-seed-development-0001-store-final context:development
insert into subscription_club_offers (name,day_count,price_credits,is_vip,is_deal,order_num) values
    ('hc_31_days',31,25,false,false,1),
    ('hc_90_days',90,65,false,true,2),
    ('vip_31_days',31,39,true,false,3),
    ('vip_90_days',90,99,true,true,4)
on conflict (name) do update set day_count=excluded.day_count,price_credits=excluded.price_credits,is_vip=excluded.is_vip,is_deal=excluded.is_deal;

insert into calendar_campaigns (name,image,start_date,day_count,enabled) values
    ('dev_holiday_calendar','calendar_dev.png',current_date,5,true)
on conflict (name) do update set start_date=excluded.start_date,day_count=excluded.day_count,enabled=true;

insert into calendar_campaign_days (campaign_id,day_number,product_definition_id,credits_reward)
select id,day,null,5 from calendar_campaigns cross join generate_series(0,4) as day where name='dev_holiday_calendar'
on conflict (campaign_id,day_number) do update set credits_reward=excluded.credits_reward;

insert into calendar_seasonal_offers (offer_date,catalog_page_id,catalog_item_id) values (current_date,2,1)
on conflict (offer_date) do update set catalog_page_id=excluded.catalog_page_id,catalog_item_id=excluded.catalog_item_id;

insert into subscription_targeted_offers (catalog_item_id,price_credits,purchase_limit,title_key,description_key,image_url,icon_url,enabled,order_num)
select 3,8,3,'offer.targeted.dev.title','offer.targeted.dev.description','','',true,1
where not exists (select 1 from subscription_targeted_offers where title_key='offer.targeted.dev.title');
--rollback delete from subscription_targeted_offer_progress where offer_id in (select id from subscription_targeted_offers where title_key='offer.targeted.dev.title'); delete from subscription_targeted_offers where title_key='offer.targeted.dev.title'; delete from calendar_seasonal_offers where offer_date=current_date; delete from calendar_door_claims where campaign_id in (select id from calendar_campaigns where name='dev_holiday_calendar'); delete from calendar_campaign_days where campaign_id in (select id from calendar_campaigns where name='dev_holiday_calendar'); delete from calendar_campaigns where name='dev_holiday_calendar'; delete from subscription_club_offers where name in ('hc_31_days','hc_90_days','vip_31_days','vip_90_days');
