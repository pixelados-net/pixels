--liquibase formatted sql

--changeset pixels:pixels-room-0006-reconcile-room-vote-scores
update rooms as room
set score = votes.score,
    updated_at = now(),
    version = room.version + 1
from (
    select candidate.id, count(room_votes.player_id)::integer as score
    from rooms as candidate
    left join room_votes on room_votes.room_id = candidate.id
    group by candidate.id
) as votes
where room.id = votes.id
  and room.score <> votes.score;

--rollback not required
