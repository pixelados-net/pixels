--liquibase formatted sql

--changeset pixels:pixels-pet-0007-stationary-plants
delete from pet_species_commands
where type_id=16 and command_id in (14,43);

--rollback insert into pet_species_commands(type_id,command_id) values (16,14),(16,43) on conflict do nothing;
