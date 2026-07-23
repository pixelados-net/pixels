--liquibase formatted sql

--changeset pixels:pixels-room-seed-development-0012-rebuild-wired-labs context:development
-- WIRED laboratories remain searchable QA rooms but are not promoted as staff picks or lifted rooms.
update rooms
set name = case id
        when 110 then 'WIRED QA Actor'
        when 111 then 'WIRED QA Conditions'
        when 112 then 'WIRED QA Movement'
        when 113 then 'WIRED QA Game'
        when 114 then 'WIRED QA Bots'
        when 115 then 'WIRED QA Safety'
    end,
    description = case id
        when 110 then 'Executable actor, badge, group, effect, hand-item, count and date conditions.'
        when 111 then 'Executable furniture, occupancy, snapshot, state, time and selector conditions.'
        when 112 then 'Walk, collision, valid-move, furniture motion, state and teleport laboratory.'
        when 113 then 'Teams, scores, lifecycle, outcome triggers, blobs and durable highscores.'
        when 114 then 'All bot actions plus reached-furniture and reached-avatar triggers.'
        when 115 then 'Durable rewards, player effects, delays, moderation and protected-target checks.'
    end,
    staff_picked = false,
    updated_at = now(),
    version = version + 1
where id between 110 and 115;

delete from navigator_lifted_rooms where room_id between 110 and 115;

--rollback update rooms set staff_picked=true where id between 110 and 115;
