package vacancy

import (
	"testing"
	"time"
)

func TestSnapshotUsesAvailableTodayHeadline(t *testing.T) {
	store := NewStore(testSettings())
	store.UpdateLoad(LoadResult{AnyRoomAvailable: true, FirstAvailableDate: "2026-04-03", FetchedAt: time.Now()})
	view := store.Snapshot(time.Date(2026, 4, 3, 10, 0, 0, 0, time.Local))
	if view.Headline != "VANDAAG KAMERS VRIJ" {
		t.Fatalf("unexpected headline: %s", view.Headline)
	}
}

func TestSnapshotUsesOverrideAsVirtualToday(t *testing.T) {
	store := NewStore(testSettings())
	store.settings.TodayDateOverride = "2026-04-10"
	store.UpdateLoad(LoadResult{AnyRoomAvailable: true, FirstAvailableDate: "2026-04-10", FetchedAt: time.Now()})
	view := store.Snapshot(time.Date(2026, 4, 3, 10, 0, 0, 0, time.Local))
	if view.Headline != "VANDAAG KAMERS VRIJ" {
		t.Fatalf("unexpected headline: %s", view.Headline)
	}
}

func TestSnapshotUsesSublineForFutureAvailability(t *testing.T) {
	store := NewStore(testSettings())
	store.UpdateLoad(LoadResult{AnyRoomAvailable: true, FirstAvailableDate: "2026-04-05", FetchedAt: time.Now()})
	view := store.Snapshot(time.Date(2026, 4, 3, 10, 0, 0, 0, time.Local))
	if !view.ShowSubline {
		t.Fatal("expected subline to be visible")
	}
	if view.Subline != "vanaf 5 april" {
		t.Fatalf("unexpected subline: %s", view.Subline)
	}
}

func TestSnapshotShowsErrorHintOnFailedAutoFetch(t *testing.T) {
	store := NewStore(testSettings())
	store.UpdateLoad(LoadResult{AnyRoomAvailable: true, FetchFailed: true, FetchedAt: time.Now()})
	view := store.Snapshot(time.Date(2026, 4, 3, 21, 0, 0, 0, time.Local))
	if !view.ShowErrorHint {
		t.Fatal("expected error hint to be visible")
	}
	if !view.Dark {
		t.Fatal("expected dark theme outside light window")
	}
}

func testSettings() Settings {
	return Settings{
		HotelName:         "Your Hotel",
		TodayDateOverride: "2026-04-03",
		DayCount:          2,
		OverrideMode:      OverrideAuto,
		ThemeMode:         ThemeTime,
		LightWindow:       ThemeWindow{StartMinute: 7 * 60, EndMinute: 19 * 60},
		PhraseRotate:      5 * time.Second,
		HeadlineScale:     80,
		ShowErrorHint:     true,
		Languages: []Language{
			{Available: "KAMERS VRIJ", AvailableToday: "VANDAAG KAMERS VRIJ", Full: "VOL", FromPrefix: "vanaf", DateLocale: "nl"},
		},
	}
}
