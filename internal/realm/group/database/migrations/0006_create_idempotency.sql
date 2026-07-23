--liquibase formatted sql

--changeset pixels:pixels-group-0006-create-idempotency
create table social_group_create_operations (
    operation_key text primary key,
    actor_player_id bigint not null references players(id),
    request_hash text not null,
    group_id bigint null references social_groups(id),
    created_at timestamptz not null default now(),
    completed_at timestamptz null,
    constraint social_group_create_operations_key_chk check(char_length(operation_key) between 8 and 128),
    constraint social_group_create_operations_hash_chk check(char_length(request_hash) = 64),
    constraint social_group_create_operations_completion_chk check((group_id is null) = (completed_at is null))
);

create index social_group_create_operations_created_idx on social_group_create_operations(created_at);

--rollback drop table if exists social_group_create_operations;
