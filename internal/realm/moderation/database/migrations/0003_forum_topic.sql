--liquibase formatted sql

--changeset pixels:pixels-moderation-0003-forum-topic
insert into cfh_topics(category,name_key,action,auto_reply_key,default_sanction_ladder,order_num,enabled)
select 'forum','moderation.topic.forum','queue',null,false,50,true
where not exists(select 1 from cfh_topics where category='forum');

--rollback delete from cfh_topics where category='forum' and name_key='moderation.topic.forum';
