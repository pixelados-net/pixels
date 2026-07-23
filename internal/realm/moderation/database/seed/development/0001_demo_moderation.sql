--liquibase formatted sql

--changeset pixels:pixels-moderation-seed-development-0001-demo context:development
--validCheckSum:ANY
insert into moderation_issues(reporter_player_id,reported_player_id,room_id,topic_id,kind,message,state)
select reporter.id,reported.id,null,topic.id,'cfh','Keeps messaging me after being asked to stop.','open'
from players reporter
join players reported on lower(reported.username)='wren'
join cfh_topics topic on topic.name_key='moderation.topic.harassment'
where lower(reporter.username)='reid'
  and not exists(select 1 from moderation_issues where message='Keeps messaging me after being asked to stop.');

insert into issue_chatlog(issue_id,player_id,pattern_id,message,created_at)
select issue.id,reported.id,'talk','ur so annoying just leave already',issue.created_at
from moderation_issues issue
join players reported on lower(reported.username)='wren'
where issue.message='Keeps messaging me after being asked to stop.'
  and not exists(select 1 from issue_chatlog where issue_id=issue.id);

insert into punishments(receiver_player_id,issuer_player_id,issuer_kind,kind,reason,source)
select target.id,issuer.id,'player','warn','Warned for disrespectful conduct in Town Hall.','admin_http'
from players target
join players issuer on lower(issuer.username)='milo'
where lower(target.username)='reid'
  and not exists(select 1 from punishments where reason='Warned for disrespectful conduct in Town Hall.');

--rollback delete from punishments where reason='Warned for disrespectful conduct in Town Hall.'; delete from moderation_issues where message='Keeps messaging me after being asked to stop.';
