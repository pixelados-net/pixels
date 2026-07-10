package catalog

import "github.com/niflaot/pixels/internal/permission"

var (
	// AdminManage allows administrative catalog management.
	AdminManage = permission.RegisterNode("catalog.admin.manage", "")
)
