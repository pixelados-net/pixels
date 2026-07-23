// Package policy declares Game Center and poll administration permissions.
package policy

import "github.com/niflaot/pixels/internal/permission"

var (
	// ManageCenter authorizes external Game Center administration.
	ManageCenter = permission.RegisterNode("games.center.manage.any", "")
	// ManagePolls authorizes room poll administration.
	ManagePolls = permission.RegisterNode("games.polls.manage.any", "")
)
