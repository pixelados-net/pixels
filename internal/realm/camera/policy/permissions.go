// Package policy defines camera permission nodes.
package policy

import "github.com/niflaot/pixels/internal/permission"

var (
	// CaptureUse permits room photo and thumbnail capture.
	CaptureUse = permission.RegisterNode("camera.capture.use", "")
	// SettingsManage permits administrative camera settings changes.
	SettingsManage = permission.RegisterNode("camera.settings.manage.any", "")
	// GalleryModerate permits gallery and capture moderation.
	GalleryModerate = permission.RegisterNode("camera.gallery.moderate.any", "")
)
