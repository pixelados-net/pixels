--liquibase formatted sql

--changeset pixels:pixels-pet-0006-contextual-training
update pet_commands
set experience_reward=0,
    energy_cost=0,
    happiness_cost=0
where id in (14,43);

--rollback update pet_commands set experience_reward=5 where id in (14,43);
