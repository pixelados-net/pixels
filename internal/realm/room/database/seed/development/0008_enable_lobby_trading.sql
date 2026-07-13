--liquibase formatted sql

--changeset pixels:pixels-room-seed-development-0008-enable-lobby-trading context:development
update rooms
set trade_mode = 2,
    updated_at = now(),
    version = version + 1
where id = 1
  and trade_mode = 0;
--rollback update rooms set trade_mode = 0, updated_at = now(), version = version + 1 where id = 1 and trade_mode = 2;
