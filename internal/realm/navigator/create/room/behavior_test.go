package create

import (
	"context"
	"errors"
	"testing"

	"github.com/niflaot/pixels/internal/command"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	netconn "github.com/niflaot/pixels/networking/connection"
	outcancreate "github.com/niflaot/pixels/networking/outbound/navigator/create/cancreate"
	outcreated "github.com/niflaot/pixels/networking/outbound/navigator/create/roomcreated"
	outalert "github.com/niflaot/pixels/networking/outbound/session/alert"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/i18n"
	"go.uber.org/zap"
)

// publisherForTest records room creation events.
type publisherForTest struct {
	// events stores published events.
	events []bus.Event
}

// Publish records one event.
func (publisher *publisherForTest) Publish(_ context.Context, event bus.Event) error {
	publisher.events = append(publisher.events, event)
	return nil
}

// TestHandleCreatesRoom verifies successful room creation projection.
func TestHandleCreatesRoom(t *testing.T) {
	players, bindings, connection, packets := commandFixture(t)
	events := &publisherForTest{}
	handler := Handler{Players: players, Bindings: bindings, Rooms: &managerForTest{created: createdRoomForTest()}, Events: events}
	err := handler.Handle(context.Background(), command.Envelope[Command]{Command: validCommand(connection)})
	if err != nil || len(*packets) != 1 || (*packets)[0].Header != outcreated.Header {
		t.Fatalf("packets=%+v err=%v", *packets, err)
	}
	if len(events.events) != 1 {
		t.Fatalf("events=%+v", events.events)
	}
}

// TestHandleReportsLimitWithoutCreating verifies native ownership-limit feedback.
func TestHandleReportsLimitWithoutCreating(t *testing.T) {
	players, bindings, connection, packets := commandFixture(t)
	rooms := make([]roommodel.Room, roomservice.MaxRoomsPerPlayer)
	handler := Handler{Players: players, Bindings: bindings, Rooms: &managerForTest{rooms: rooms}}
	if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: validCommand(connection)}); err != nil {
		t.Fatalf("handle limit: %v", err)
	}
	if len(*packets) != 1 || (*packets)[0].Header != outcancreate.Header {
		t.Fatalf("packets=%+v", *packets)
	}
}

// TestHandleReportsLocalizedValidationError verifies soft creation failures.
func TestHandleReportsLocalizedValidationError(t *testing.T) {
	players, bindings, connection, packets := commandFixture(t)
	translations := i18n.NewCatalog(i18n.Config{}, map[i18n.Locale]map[i18n.Key]string{"en": {"navigator.create.error.layout_unavailable": "Layout unavailable"}})
	handler := Handler{Players: players, Bindings: bindings, Rooms: &managerForTest{createErr: roomservice.ErrLayoutNotAvailable}, Translations: translations, Log: zap.NewNop()}
	if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: validCommand(connection)}); err != nil {
		t.Fatalf("handle validation: %v", err)
	}
	if len(*packets) != 1 || (*packets)[0].Header != outalert.Header {
		t.Fatalf("packets=%+v", *packets)
	}
}

// TestHandlePropagatesPersistenceFailures verifies unexpected errors remain hard failures.
func TestHandlePropagatesPersistenceFailures(t *testing.T) {
	players, bindings, connection, _ := commandFixture(t)
	cause := errors.New("database unavailable")
	handler := Handler{Players: players, Bindings: bindings, Rooms: &managerForTest{listErr: cause}}
	err := handler.Handle(context.Background(), command.Envelope[Command]{Command: validCommand(connection)})
	if !errors.Is(err, cause) {
		t.Fatalf("error=%v", err)
	}
}

// TestCreateErrorKeyMapsExpectedFailures verifies localized validation families.
func TestCreateErrorKeyMapsExpectedFailures(t *testing.T) {
	errorsToMap := []error{
		roomservice.ErrLayoutNotAvailable, roomservice.ErrInvalidCategory,
		roomservice.ErrInvalidRoomName, roomservice.ErrProhibitedName,
		roomservice.ErrInvalidDescription, roomservice.ErrProhibitedDescription,
		roomservice.ErrInvalidMaxUsers, roomservice.ErrInvalidTradeMode,
	}
	for _, err := range errorsToMap {
		if key, soft := createErrorKey(err); !soft || key == "" {
			t.Fatalf("error=%v key=%q soft=%v", err, key, soft)
		}
	}
	if _, soft := createErrorKey(errors.New("unexpected")); soft {
		t.Fatal("unexpected error must remain hard")
	}
}

// validCommand creates one valid room creation command.
func validCommand(connection netconn.Context) Command {
	return Command{Handler: connection, RoomName: "Created Room", ModelName: "model_a", CategoryID: 1, MaxVisitors: 25}
}
