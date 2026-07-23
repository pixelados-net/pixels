package currency

import "github.com/niflaot/pixels/internal/permission"

var (
	// AdminManage allows administrative currency management.
	AdminManage = permission.RegisterNode("currency.admin.manage", "")
	// InfiniteBalance exempts player-originated deductions from balance checks.
	InfiniteBalance = permission.RegisterNode("currency.economy.infinite", "")
)
