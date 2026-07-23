package guide

import (
	"strings"
	"testing"
	"time"
)

// TestSessionLifecycle verifies FIFO matching, rejection rematch, chat, and feedback state.
func TestSessionLifecycle(t *testing.T) {
	manager := New(nil)
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	manager.now = func() time.Time { return now }
	manager.SetDuty(2, true, false, false)
	now = now.Add(time.Second)
	manager.SetDuty(3, true, false, false)
	session, err := manager.Create(1, 7, "  need   help ")
	if err != nil || session.GuidePlayerID != 2 || session.Description != "need help" {
		t.Fatalf("create session=%+v err=%v", session, err)
	}
	session, err = manager.Decide(2, false)
	if err != nil || session.GuidePlayerID != 3 || session.State != StateAttached {
		t.Fatalf("rematch session=%+v err=%v", session, err)
	}
	if _, err = manager.Decide(3, true); err != nil {
		t.Fatal(err)
	}
	_, message, err := manager.Send(1, " hello   guide ")
	if err != nil || message.Text != "hello guide" {
		t.Fatalf("message=%+v err=%v", message, err)
	}
	ended, found := manager.End(3)
	if !found || ended.State != StateEnded {
		t.Fatalf("ended=%+v found=%v", ended, found)
	}
	completed, found := manager.TakeCompleted(1)
	if !found || completed.GuidePlayerID != 3 || len(completed.Transcript) != 1 {
		t.Fatalf("completed=%+v found=%v", completed, found)
	}
}

// TestSessionErrorsAndTranscriptBound verifies invalid actors, states, and memory bounds.
func TestSessionErrorsAndTranscriptBound(t *testing.T) {
	manager := New(nil)
	manager.SetDuty(2, true, false, false)
	if _, err := manager.Create(1, 1, "help"); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Create(1, 1, "again"); err != ErrBusy {
		t.Fatalf("busy err=%v", err)
	}
	if _, err := manager.Decide(3, true); err != ErrUnauthorized {
		t.Fatalf("unauthorized err=%v", err)
	}
	if _, _, err := manager.Send(1, "early"); err != ErrInvalidState {
		t.Fatalf("early send err=%v", err)
	}
	if _, err := manager.Decide(2, true); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Decide(2, true); err != ErrInvalidState {
		t.Fatalf("second decision err=%v", err)
	}
	for index := 0; index < 205; index++ {
		if _, _, err := manager.Send(1, "message"); err != nil {
			t.Fatal(err)
		}
	}
	session, found := manager.SessionFor(1)
	if !found || len(session.Transcript) != 200 {
		t.Fatalf("found=%v messages=%d", found, len(session.Transcript))
	}
	_, message, err := manager.Send(1, strings.Repeat("x", 600))
	if err != nil || len(message.Text) != 500 {
		t.Fatalf("length=%d err=%v", len(message.Text), err)
	}
}

// TestDutyPoolsExcludeBusyPlayers verifies counts and guardian FIFO selection.
func TestDutyPoolsExcludeBusyPlayers(t *testing.T) {
	manager := New(nil)
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	manager.now = func() time.Time { return now }
	manager.SetDuty(2, true, true, true)
	now = now.Add(time.Second)
	manager.SetDuty(3, true, false, true)
	guides, bullies, guardians := manager.DutyCount()
	if guides != 2 || bullies != 1 || guardians != 2 {
		t.Fatalf("counts=%d/%d/%d", guides, bullies, guardians)
	}
	if ids := manager.Guardians(0, 1); len(ids) != 1 || ids[0] != 2 {
		t.Fatalf("guardians=%v", ids)
	}
	if _, err := manager.Create(1, 1, "help"); err != nil {
		t.Fatal(err)
	}
	guides, bullies, guardians = manager.DutyCount()
	if guides != 1 || bullies != 0 || guardians != 1 {
		t.Fatalf("busy counts=%d/%d/%d", guides, bullies, guardians)
	}
	manager.SetDuty(3, false, false, false)
	if ids := manager.Guardians(0, 5); len(ids) != 0 {
		t.Fatalf("guardians=%v", ids)
	}
}

// TestRequesterIsNeverRematchedAsGuide verifies rejection cannot self-pair a duty requester.
func TestRequesterIsNeverRematchedAsGuide(t *testing.T) {
	manager := New(nil)
	manager.SetDuty(1, true, false, false)
	manager.SetDuty(2, true, false, false)
	if _, err := manager.Create(1, 1, "help"); err != nil {
		t.Fatal(err)
	}
	session, err := manager.Decide(2, false)
	if err != ErrUnavailable || session.GuidePlayerID != 2 {
		t.Fatalf("session=%+v err=%v", session, err)
	}
}

// TestDisconnectReleasesSession verifies indexes and duty are removed together.
func TestDisconnectReleasesSession(t *testing.T) {
	manager := New(nil)
	manager.SetDuty(2, true, false, false)
	if _, err := manager.Create(1, 1, "help"); err != nil {
		t.Fatal(err)
	}
	session, found := manager.RemovePlayer(2)
	if !found || session.RequesterPlayerID != 1 {
		t.Fatalf("session=%+v found=%v", session, found)
	}
	if _, found = manager.SessionFor(1); found {
		t.Fatal("requester remained indexed after guide disconnect")
	}
}
