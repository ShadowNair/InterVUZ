package domain

import (
	"strings"
	"testing"
)

func TestSyntheticClassroomFromPlaceID(t *testing.T) {
	place, ok := SyntheticClassroomFromPlaceID("place_223")
	if !ok {
		t.Fatal("expected synthetic classroom")
	}
	if place.ID != "place_223" {
		t.Fatalf("unexpected id: %s", place.ID)
	}
	if NormalizeRoomName(place.Name) != "223" {
		t.Fatalf("unexpected normalized room: %s", NormalizeRoomName(place.Name))
	}
	if place.Type != "classroom" {
		t.Fatalf("unexpected type: %s", place.Type)
	}

	composite, ok := SyntheticClassroomFromPlaceID("place_225_219")
	if !ok {
		t.Fatal("expected composite synthetic classroom")
	}
	if !strings.Contains(composite.Name, "225") || !strings.Contains(composite.Name, "219") {
		t.Fatalf("unexpected composite name: %s", composite.Name)
	}

	if _, ok := SyntheticClassroomFromPlaceID("room-223"); ok {
		t.Fatal("expected non-place id to be rejected")
	}
}
