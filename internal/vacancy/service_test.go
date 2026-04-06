package vacancy

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeLookup struct {
	dates map[int]string
	err   error
}

func (f fakeLookup) LookupFirstArrival(
	ctx context.Context,
	roomID int,
	from string,
	until string,
	targets map[string]struct{},
) (string, error) {
	_ = ctx
	_ = from
	_ = until
	_ = targets
	if f.err != nil {
		return "", f.err
	}
	return f.dates[roomID], nil
}

func TestCheckInAvailabilityReturnsEarliestDate(t *testing.T) {
	service := NewService(fakeLookup{dates: map[int]string{1: "2026-04-05", 2: "2026-04-04"}}, []int{1, 2}, 2)
	result, err := service.CheckInAvailability(context.Background(), time.Date(2026, 4, 4, 0, 0, 0, 0, time.Local), 2)
	if err != nil {
		t.Fatalf("CheckInAvailability returned error: %v", err)
	}
	assertAvailability(t, result)
}

func TestCheckInAvailabilityReturnsError(t *testing.T) {
	service := NewService(fakeLookup{err: errors.New("boom")}, []int{1}, 1)
	_, err := service.CheckInAvailability(context.Background(), time.Now(), 2)
	if err == nil {
		t.Fatal("CheckInAvailability should return an error")
	}
}

func assertAvailability(t *testing.T, result CheckInAvailability) {
	t.Helper()
	if !result.AnyRoomAvailable {
		t.Fatal("expected AnyRoomAvailable to be true")
	}
	if result.FirstAvailableDate != "2026-04-04" {
		t.Fatalf("unexpected FirstAvailableDate: %s", result.FirstAvailableDate)
	}
}
