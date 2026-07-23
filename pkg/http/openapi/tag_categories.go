package openapi

import "strings"

// operationCategory refines broad route families into focused OpenAPI sections.
func operationCategory(path string, fallback string) string {
	switch fallback {
	case "Players":
		return playerCategory(path)
	case "Pets":
		return petCategory(path)
	case "Camera":
		return cameraCategory(path)
	case "Progression":
		return progressionCategory(path)
	case "Games":
		return gamesCategory(path)
	case "Groups":
		return groupCategory(path)
	case "Crafting":
		return craftingCategory(path)
	case "Catalog":
		return catalogCategory(path)
	case "Subscriptions":
		return subscriptionCategory(path)
	case "Chat":
		return chatCategory(path)
	case "Rooms":
		return roomCategory(path)
	case "Trading":
		return tradingCategory(path)
	case "Moderation":
		return moderationCategory(path)
	default:
		return fallback
	}
}

// playerCategory separates effect inventory from general player administration.
func playerCategory(path string) string {
	if strings.Contains(path, "/effects") {
		return "Player Effects"
	}

	return "Players"
}

// petCategory separates reference data from individual pet operations.
func petCategory(path string) string {
	for _, segment := range []string{"/species", "/breeds", "/commands", "/reference/"} {
		if strings.Contains(path, segment) {
			return "Pet Reference"
		}
	}

	return "Pets"
}

// cameraCategory separates published photos from capture configuration.
func cameraCategory(path string) string {
	if strings.Contains(path, "/gallery") || strings.Contains(path, "/photos/") {
		return "Photo Gallery"
	}

	return "Camera"
}

// progressionCategory separates the independent progression systems.
func progressionCategory(path string) string {
	switch {
	case strings.Contains(path, "/achievements"), strings.Contains(path, "/badges"):
		return "Achievements"
	case strings.Contains(path, "/talents"):
		return "Talents"
	case strings.Contains(path, "/campaigns"), strings.Contains(path, "/quests"):
		return "Quests"
	case strings.Contains(path, "/quizzes"), strings.Contains(path, "/polls"):
		return "Quizzes"
	case strings.Contains(path, "/promos"):
		return "Badge Promotions"
	default:
		return "Progression"
	}
}

// gamesCategory separates external games, room polls, and room score data.
func gamesCategory(path string) string {
	switch {
	case strings.Contains(path, "/center"):
		return "Game Center"
	case strings.Contains(path, "/polls"):
		return "Room Polls"
	default:
		return "Room Games"
	}
}

// groupCategory separates membership and forum workflows from group records.
func groupCategory(path string) string {
	switch {
	case strings.Contains(path, "/forum/"):
		return "Group Forums"
	case strings.Contains(path, "/members"), strings.Contains(path, "/requests"), strings.Contains(path, "/groups/players"):
		return "Group Members"
	default:
		return "Groups"
	}
}

// craftingCategory separates recycler configuration from crafting recipes.
func craftingCategory(path string) string {
	if strings.Contains(path, "/recycler/") {
		return "Recycler"
	}

	return "Crafting"
}

// catalogCategory separates pages, offers, and vouchers.
func catalogCategory(path string) string {
	switch {
	case strings.Contains(path, "/vouchers"):
		return "Catalog Vouchers"
	case strings.Contains(path, "/items"), strings.Contains(path, "/sanitize-list"):
		return "Catalog Offers"
	default:
		return "Catalog Pages"
	}
}

// subscriptionCategory separates membership and commercial offer workflows.
func subscriptionCategory(path string) string {
	switch {
	case strings.Contains(path, "/club-offers"):
		return "Club Offers"
	case strings.Contains(path, "/targeted-offers"):
		return "Targeted Offers"
	case strings.Contains(path, "/calendar/"):
		return "Calendar"
	default:
		return "Memberships"
	}
}

// chatCategory separates configuration from audit history.
func chatCategory(path string) string {
	switch {
	case strings.Contains(path, "/filters"):
		return "Chat Filters"
	case strings.Contains(path, "/bubbles"):
		return "Chat Bubbles"
	default:
		return "Chat History"
	}
}

// roomCategory separates persistent room data from live runtime controls.
func roomCategory(path string) string {
	switch {
	case strings.Contains(path, "/promotion"):
		return "Room Promotions"
	case strings.Contains(path, "/occupancy"), strings.Contains(path, "/roller"), strings.HasSuffix(path, "/close"),
		strings.HasSuffix(path, "/forward"), strings.HasSuffix(path, "/teleport"):
		return "Room Runtime"
	default:
		return "Rooms"
	}
}

// tradingCategory separates direct trades from Marketplace intervention.
func tradingCategory(path string) string {
	if strings.Contains(path, "/marketplace/") {
		return "Marketplace"
	}

	return "Player Trading"
}

// moderationCategory separates sanctions, cases, and policy configuration.
func moderationCategory(path string) string {
	switch {
	case strings.Contains(path, "/punishments"):
		return "Punishments"
	case strings.Contains(path, "/cfh-topics"), strings.Contains(path, "/presets"), strings.Contains(path, "/sanction-ladder"):
		return "Moderation Settings"
	default:
		return "Moderation"
	}
}
