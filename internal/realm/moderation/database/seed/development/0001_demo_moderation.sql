--liquibase formatted sql

--changeset pixels:pixels-moderation-seed-development-0001-demo context:development
insert into moderation_issues(reporter_player_id,reported_player_id,room_id,topic_id,kind,message,state)
select reporter.id,reported.id,null,topic.id,'cfh','Development moderation issue','open'
from players reporter
join players reported on lower(reported.username)='carol'
join cfh_topics topic on topic.name_key='moderation.topic.harassment'
where lower(reporter.username)='bob'
  and not exists(select 1 from moderation_issues where message='Development moderation issue');

insert into issue_chatlog(issue_id,player_id,pattern_id,message,created_at)
select issue.id,reported.id,'talk','Development frozen chat evidence',issue.created_at
from moderation_issues issue
join players reported on lower(reported.username)='carol'
where issue.message='Development moderation issue'
  and not exists(select 1 from issue_chatlog where issue_id=issue.id);

insert into punishments(receiver_player_id,issuer_player_id,issuer_kind,kind,reason,source)
select target.id,issuer.id,'player','warn','Development warning history','admin_http'
from players target
join players issuer on lower(issuer.username)='demo'
where lower(target.username)='bob'
  and not exists(select 1 from punishments where reason='Development warning history');

--rollback delete from punishments where reason='Development warning history'; delete from moderation_issues where message='Development moderation issue';
