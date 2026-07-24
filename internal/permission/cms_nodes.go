package permission

var (
	// CMSNewsManage allows creating and editing website news.
	CMSNewsManage = RegisterNode("cms.news.manage", "")
	// CMSMaintenanceManage allows changing maintenance windows.
	CMSMaintenanceManage = RegisterNode("cms.maintenance.manage", "")
	// CMSMaintenanceBypass allows entering the website during maintenance.
	CMSMaintenanceBypass = RegisterNode("cms.maintenance.bypass", "")
	// CMSMaintenanceEarlyAccessManage allows assigning temporary maintenance access.
	CMSMaintenanceEarlyAccessManage = RegisterNode("cms.maintenance.early_access.manage", "")
	// CMSStorePackagesManage allows changing store packages.
	CMSStorePackagesManage = RegisterNode("cms.store.packages.manage", "")
	// CMSStoreTransactionsView allows reading store transactions.
	CMSStoreTransactionsView = RegisterNode("cms.store.transactions.view", "")
	// CMSStoreTransactionsAuthorize allows authorizing pending store transactions.
	CMSStoreTransactionsAuthorize = RegisterNode("cms.store.transactions.authorize", "")
	// CMSPermissionGroupsView allows reading permission group administration.
	CMSPermissionGroupsView = RegisterNode("cms.permissions.groups.view", "")
	// CMSPermissionGroupsCreate allows creating permission groups.
	CMSPermissionGroupsCreate = RegisterNode("cms.permissions.groups.create", "")
	// CMSPermissionGroupsUpdate allows editing permission group metadata.
	CMSPermissionGroupsUpdate = RegisterNode("cms.permissions.groups.update", "")
	// CMSPermissionGroupNodesManage allows changing group permission grants.
	CMSPermissionGroupNodesManage = RegisterNode("cms.permissions.groups.nodes.manage", "")
	// CMSPermissionGroupMembersManage allows changing direct group memberships.
	CMSPermissionGroupMembersManage = RegisterNode("cms.permissions.groups.members.manage", "")
)
