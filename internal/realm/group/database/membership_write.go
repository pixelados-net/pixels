package database

import (
	"context"

	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
)

// Join inserts a member or pending request and reports pending and changed state.
func (repository *Repository) Join(ctx context.Context, groupID int64, playerID int64, membershipLimit int, memberLimit int, pendingLimit int) (grouprecord.Membership, bool, bool, error) {
	var member grouprecord.Membership
	var pending, changed bool
	err := repository.WithinTransaction(ctx, func(txCtx context.Context) error {
		var state grouprecord.State
		var members, requests int
		if err := repository.executor(txCtx).QueryRow(txCtx, `select state,member_count,pending_count from social_groups where id=$1 and deactivated_at is null for update`, groupID).Scan(&state, &members, &requests); err != nil {
			return grouprecord.ErrNotFound
		}
		if existing, found, err := repository.Membership(txCtx, groupID, playerID); err != nil {
			return err
		} else if found {
			member = existing
			return nil
		}
		var exists bool
		if err := repository.executor(txCtx).QueryRow(txCtx, `select exists(select 1 from social_group_requests where group_id=$1 and player_id=$2)`, groupID, playerID).Scan(&exists); err != nil {
			return err
		}
		if exists {
			pending = true
			return nil
		}
		if state == grouprecord.Private {
			return grouprecord.ErrClosed
		}
		count, err := repository.CountMemberships(txCtx, playerID)
		if err != nil {
			return err
		}
		if count >= membershipLimit {
			return grouprecord.ErrLimit
		}
		if state == grouprecord.Exclusive {
			if requests >= pendingLimit {
				return grouprecord.ErrLimit
			}
			if _, err = repository.executor(txCtx).Exec(txCtx, `insert into social_group_requests(group_id,player_id) values($1,$2)`, groupID, playerID); err != nil {
				return mapConflict(err)
			}
			_, err = repository.executor(txCtx).Exec(txCtx, `update social_groups set pending_count=pending_count+1,updated_at=now(),version=version+1 where id=$1`, groupID)
			pending = true
			changed = err == nil
			return err
		}
		if members >= memberLimit {
			return grouprecord.ErrLimit
		}
		if _, err = repository.executor(txCtx).Exec(txCtx, `insert into social_group_members(group_id,player_id,role) values($1,$2,2)`, groupID, playerID); err != nil {
			return mapConflict(err)
		}
		if _, err = repository.executor(txCtx).Exec(txCtx, `update social_groups set member_count=member_count+1,updated_at=now(),version=version+1 where id=$1`, groupID); err != nil {
			return err
		}
		member, _, err = repository.Membership(txCtx, groupID, playerID)
		changed = err == nil
		return err
	})
	return member, pending, changed, err
}

// AddMember administratively inserts or replaces one non-owner role.
func (repository *Repository) AddMember(ctx context.Context, groupID int64, playerID int64, role grouprecord.Role, membershipLimit int, memberLimit int) (grouprecord.Membership, bool, error) {
	if role != grouprecord.Admin && role != grouprecord.Member {
		return grouprecord.Membership{}, false, grouprecord.ErrInvalid
	}
	created := false
	var member grouprecord.Membership
	err := repository.WithinTransaction(ctx, func(txCtx context.Context) error {
		var members int
		if err := repository.executor(txCtx).QueryRow(txCtx, `select member_count from social_groups where id=$1 and deactivated_at is null for update`, groupID).Scan(&members); err != nil {
			return grouprecord.ErrNotFound
		}
		existing, found, err := repository.Membership(txCtx, groupID, playerID)
		if err != nil {
			return err
		}
		if found {
			if existing.Role == grouprecord.Owner {
				return grouprecord.ErrForbidden
			}
			member, err = repository.ChangeRole(txCtx, groupID, playerID, role)
			return err
		}
		if members >= memberLimit {
			return grouprecord.ErrLimit
		}
		count, err := repository.CountMemberships(txCtx, playerID)
		if err != nil {
			return err
		}
		if count >= membershipLimit {
			return grouprecord.ErrLimit
		}
		command, err := repository.executor(txCtx).Exec(txCtx, `delete from social_group_requests where group_id=$1 and player_id=$2`, groupID, playerID)
		if err != nil {
			return err
		}
		if _, err = repository.executor(txCtx).Exec(txCtx, `insert into social_group_members(group_id,player_id,role) values($1,$2,$3)`, groupID, playerID, role); err != nil {
			return mapConflict(err)
		}
		if _, err = repository.executor(txCtx).Exec(txCtx, `update social_groups set member_count=member_count+1,pending_count=greatest(pending_count-$2,0),updated_at=now(),version=version+1 where id=$1`, groupID, command.RowsAffected()); err != nil {
			return err
		}
		created = true
		member, _, err = repository.Membership(txCtx, groupID, playerID)
		if err == nil {
			err = repository.audit(txCtx, groupID, "group.member.added", playerID, member.Version)
		}
		return err
	})
	return member, created, err
}

// AcceptRequest promotes one locked request to membership.
func (repository *Repository) AcceptRequest(ctx context.Context, groupID int64, playerID int64, memberLimit int) (grouprecord.Membership, error) {
	var member grouprecord.Membership
	err := repository.WithinTransaction(ctx, func(txCtx context.Context) error {
		var count int
		if err := repository.executor(txCtx).QueryRow(txCtx, `select member_count from social_groups where id=$1 and deactivated_at is null for update`, groupID).Scan(&count); err != nil {
			return grouprecord.ErrNotFound
		}
		if count >= memberLimit {
			return grouprecord.ErrLimit
		}
		command, err := repository.executor(txCtx).Exec(txCtx, `delete from social_group_requests where group_id=$1 and player_id=$2`, groupID, playerID)
		if err != nil {
			return err
		}
		if command.RowsAffected() != 1 {
			return grouprecord.ErrNotFound
		}
		if _, err = repository.executor(txCtx).Exec(txCtx, `insert into social_group_members(group_id,player_id,role) values($1,$2,2) on conflict do nothing`, groupID, playerID); err != nil {
			return mapConflict(err)
		}
		if _, err = repository.executor(txCtx).Exec(txCtx, `update social_groups set member_count=member_count+1,pending_count=greatest(pending_count-1,0),updated_at=now(),version=version+1 where id=$1`, groupID); err != nil {
			return err
		}
		member, _, err = repository.Membership(txCtx, groupID, playerID)
		if err == nil {
			err = repository.audit(txCtx, groupID, "group.request.accepted", playerID, member.Version)
		}
		return err
	})
	return member, err
}

// DeclineRequest removes one request idempotently.
func (repository *Repository) DeclineRequest(ctx context.Context, groupID int64, playerID int64) (bool, error) {
	removed := false
	err := repository.WithinTransaction(ctx, func(txCtx context.Context) error {
		command, err := repository.executor(txCtx).Exec(txCtx, `delete from social_group_requests where group_id=$1 and player_id=$2`, groupID, playerID)
		if err != nil {
			return err
		}
		removed = command.RowsAffected() == 1
		if !removed {
			return nil
		}
		if _, err = repository.executor(txCtx).Exec(txCtx, `update social_groups set pending_count=greatest(pending_count-1,0),updated_at=now(),version=version+1 where id=$1`, groupID); err != nil {
			return err
		}
		return repository.audit(txCtx, groupID, "group.request.declined", playerID, 0)
	})
	return removed, err
}

// ApproveAll accepts a bounded ordered request batch.
func (repository *Repository) ApproveAll(ctx context.Context, groupID int64, limit int, memberLimit int) ([]grouprecord.Membership, error) {
	requests, err := repository.Requests(ctx, groupID, 0, limit)
	if err != nil {
		return nil, err
	}
	accepted := make([]grouprecord.Membership, 0, len(requests))
	err = repository.WithinTransaction(ctx, func(txCtx context.Context) error {
		for _, request := range requests {
			member, acceptErr := repository.AcceptRequest(txCtx, groupID, request.PlayerID, memberLimit)
			if acceptErr != nil {
				return acceptErr
			}
			accepted = append(accepted, member)
		}
		return nil
	})
	return accepted, err
}

// ChangeRole changes one target role while preserving owner invariants.
