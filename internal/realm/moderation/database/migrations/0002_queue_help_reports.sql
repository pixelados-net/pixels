--liquibase formatted sql

--changeset pixels:pixels-moderation-0002-queue-help-reports
update cfh_topics
set action = 'queue',
    auto_reply_key = null
where name_key = 'moderation.topic.help'
  and action = 'auto_reply';

--rollback update cfh_topics set action = 'auto_reply', auto_reply_key = 'moderation.reply.help' where name_key = 'moderation.topic.help';
