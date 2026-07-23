package trigger

import (
	camerapurchased "github.com/niflaot/pixels/internal/realm/camera/gallery/events/purchased"
	cataloggift "github.com/niflaot/pixels/internal/realm/catalog/events/gift"
	craftcrafted "github.com/niflaot/pixels/internal/realm/crafting/recipe/events/crafted"
	craftdiscovered "github.com/niflaot/pixels/internal/realm/crafting/recipe/events/discovered"
	craftrecycled "github.com/niflaot/pixels/internal/realm/crafting/recycler/events/recycled"
	fireworkcharged "github.com/niflaot/pixels/internal/realm/furniture/events/fireworkcharged"
	furnitureplaced "github.com/niflaot/pixels/internal/realm/furniture/events/placed"
	postitplaced "github.com/niflaot/pixels/internal/realm/furniture/events/postitplaced"
	surfaceapplied "github.com/niflaot/pixels/internal/realm/furniture/events/surfaceapplied"
	friendaccepted "github.com/niflaot/pixels/internal/realm/messenger/friend/events/requestaccepted"
	playerignored "github.com/niflaot/pixels/internal/realm/messenger/session/events/ignored"
	guideenrolled "github.com/niflaot/pixels/internal/realm/moderation/events/guideenrolled"
	guidesession "github.com/niflaot/pixels/internal/realm/moderation/events/sessioncompleted"
	planthealed "github.com/niflaot/pixels/internal/realm/pet/breeding/plant/events/healed"
	planttreated "github.com/niflaot/pixels/internal/realm/pet/breeding/plant/events/treated"
	petfed "github.com/niflaot/pixels/internal/realm/pet/care/events/fed"
	petleveled "github.com/niflaot/pixels/internal/realm/pet/care/events/leveled"
	petrespected "github.com/niflaot/pixels/internal/realm/pet/care/events/respected"
	petcreated "github.com/niflaot/pixels/internal/realm/pet/identity/events/created"
	playerauthenticated "github.com/niflaot/pixels/internal/realm/player/events/authenticated"
	playerdisconnected "github.com/niflaot/pixels/internal/realm/player/events/disconnected"
	profilerespect "github.com/niflaot/pixels/internal/realm/player/profile/events/respectgranted"
	profileupdated "github.com/niflaot/pixels/internal/realm/player/profile/events/updated"
	roomentered "github.com/niflaot/pixels/internal/realm/room/access/events/entered"
	roomkicked "github.com/niflaot/pixels/internal/realm/room/control/events/kicked"
	roommuted "github.com/niflaot/pixels/internal/realm/room/control/events/muted"
	roomsettings "github.com/niflaot/pixels/internal/realm/room/control/events/settingsupdated"
	roomwordfilter "github.com/niflaot/pixels/internal/realm/room/control/events/wordfiltermodified"
	roomstaffpicked "github.com/niflaot/pixels/internal/realm/room/record/events/staffpicked"
	gameprogressed "github.com/niflaot/pixels/internal/realm/room/world/games/events/progressed"
	subscriptionpayday "github.com/niflaot/pixels/internal/realm/subscription/events/payday"
	tradecompleted "github.com/niflaot/pixels/internal/realm/trade/events/completed"
)

// registrations returns the immutable event adapter table.
func (subscriber *Subscriber) registrations() []registration {
	return []registration{
		{playerauthenticated.Name, subscriber.authenticated},
		{playerdisconnected.Name, subscriber.disconnected},
		{roomentered.Name, subscriber.roomEntered},
		{profileupdated.Name, subscriber.profileUpdated},
		{profilerespect.Name, subscriber.respect},
		{furnitureplaced.Name, subscriber.furniturePlaced},
		{fireworkcharged.Name, subscriber.fireworkCharged},
		{surfaceapplied.Name, subscriber.surfaceApplied},
		{postitplaced.Name, subscriber.postItPlaced},
		{friendaccepted.Name, subscriber.friendAccepted},
		{playerignored.Name, subscriber.playerIgnored},
		{cataloggift.Name, subscriber.gift},
		{camerapurchased.Name, subscriber.cameraPurchased},
		{craftcrafted.Name, subscriber.crafted},
		{craftdiscovered.Name, subscriber.discovered},
		{craftrecycled.Name, subscriber.recycled},
		{petcreated.Name, subscriber.petCreated},
		{petleveled.Name, subscriber.petLeveled},
		{petrespected.Name, subscriber.petRespected},
		{petfed.Name, subscriber.petFed},
		{planttreated.Name, subscriber.plantTreated},
		{planthealed.Name, subscriber.plantHealed},
		{guideenrolled.Name, subscriber.guideEnrolled},
		{guidesession.Name, subscriber.guideSession},
		{tradecompleted.Name, subscriber.tradeCompleted},
		{subscriptionpayday.Name, subscriber.subscriptionPayday},
		{roomsettings.Name, subscriber.selfModSettings},
		{roomwordfilter.Name, subscriber.selfModWordFilter},
		{roommuted.Name, subscriber.selfModMuted},
		{roomkicked.Name, subscriber.selfModKicked},
		{roomstaffpicked.Name, subscriber.staffPicked},
		{gameprogressed.Name, subscriber.gameProgressed},
	}
}
