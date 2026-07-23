package core

import "context"

// Offer validates and atomically stages one or more furniture items.
func (service *Service) Offer(ctx context.Context, playerID int64, itemIDs []int64) error {
	session, found := service.registry.Find(playerID)
	if !found {
		return ErrUnavailable
	}
	participant, _ := session.Participant(playerID)
	if len(participant.Items)+len(itemIDs) > service.config.MaximumItems {
		return ErrMaximumItems
	}
	for _, itemID := range itemIDs {
		item, found, err := service.furniture.FindItemByID(ctx, itemID)
		if err != nil {
			return err
		}
		if !found || item.OwnerPlayerID != playerID || !item.InInventory() || item.MarketplaceReserved {
			return ErrItemUnavailable
		}
		definition, found, err := service.furniture.FindDefinitionByID(ctx, item.DefinitionID)
		if err != nil {
			return err
		}
		if !found || !definition.AllowTrade {
			return ErrItemUnavailable
		}
	}
	if !service.registry.StageMany(playerID, itemIDs) {
		return ErrItemUnavailable
	}
	return nil
}

// Remove removes one staged item before first-phase acceptance.
func (service *Service) Remove(playerID int64, itemID int64) error {
	session, found := service.registry.Find(playerID)
	if !found {
		return ErrUnavailable
	}
	participant, found := session.Participant(playerID)
	if !found {
		return ErrUnavailable
	}
	if participant.Accepted {
		return ErrAccepted
	}
	if !service.registry.Unstage(playerID, itemID) {
		return ErrItemUnavailable
	}
	return nil
}

// Accept updates first-phase agreement.
func (service *Service) Accept(playerID int64, accepted bool) (bool, error) {
	session, found := service.registry.Find(playerID)
	if !found {
		return false, ErrUnavailable
	}
	both, updated := session.SetAccepted(playerID, accepted)
	if !updated {
		return false, ErrUnavailable
	}
	return both, nil
}
