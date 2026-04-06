package vacancy

import (
	"context"
	"sort"
	"sync"
	"time"
)

type ArrivalLookup interface {
	LookupFirstArrival(
		ctx context.Context,
		roomID int,
		from string,
		until string,
		targets map[string]struct{},
	) (string, error)
}

type Service struct {
	lookup      ArrivalLookup
	roomIDs     []int
	concurrency int
}

type roomPlan struct {
	from    string
	until   string
	targets map[string]struct{}
}

type roomResult struct {
	date string
	err  error
}

const bookzoUntilHorizonExtraDays = 14

func NewService(lookup ArrivalLookup, roomIDs []int, concurrency int) *Service {
	return &Service{lookup: lookup, roomIDs: roomIDs, concurrency: maxConcurrency(concurrency)}
}

func maxConcurrency(value int) int {
	if value < 1 {
		return 1
	}
	return value
}

func (s *Service) CheckInAvailability(
	ctx context.Context,
	refDate time.Time,
	dayCount int,
) (CheckInAvailability, error) {
	plan := newRoomPlan(refDate, dayCount)
	results := s.collectResults(ctx, plan)
	return reduceResults(results)
}

func newRoomPlan(refDate time.Time, dayCount int) roomPlan {
	until := AddDays(refDate, dayCount+bookzoUntilHorizonExtraDays)
	return roomPlan{
		from:    FormatYMD(refDate),
		until:   FormatYMD(until),
		targets: TargetDateSet(refDate, dayCount),
	}
}

func (s *Service) collectResults(ctx context.Context, plan roomPlan) <-chan roomResult {
	results := make(chan roomResult, len(s.roomIDs))
	go s.runChecks(ctx, plan, results)
	return results
}

func (s *Service) runChecks(ctx context.Context, plan roomPlan, results chan<- roomResult) {
	defer close(results)
	wg := &sync.WaitGroup{}
	sem := make(chan struct{}, s.concurrency)
	for _, roomID := range s.roomIDs {
		s.startRoomCheck(ctx, wg, sem, results, roomID, plan)
	}
	wg.Wait()
}

func (s *Service) startRoomCheck(
	ctx context.Context,
	wg *sync.WaitGroup,
	sem chan struct{},
	results chan<- roomResult,
	roomID int,
	plan roomPlan,
) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		sem <- struct{}{}
		defer func() { <-sem }()
		results <- s.lookupRoom(ctx, roomID, plan)
	}()
}

func (s *Service) lookupRoom(ctx context.Context, roomID int, plan roomPlan) roomResult {
	date, err := s.lookup.LookupFirstArrival(ctx, roomID, plan.from, plan.until, plan.targets)
	return roomResult{date: date, err: err}
}

func reduceResults(results <-chan roomResult) (CheckInAvailability, error) {
	dates := []string{}
	for result := range results {
		if result.err != nil {
			return CheckInAvailability{}, result.err
		}
		dates = appendIfDate(dates, result.date)
	}
	return availabilityFromDates(dates), nil
}

func appendIfDate(dates []string, value string) []string {
	if value == "" {
		return dates
	}
	return append(dates, value)
}

func availabilityFromDates(dates []string) CheckInAvailability {
	sort.Strings(dates)
	return CheckInAvailability{
		AnyRoomAvailable:   len(dates) > 0,
		FirstAvailableDate: firstDate(dates),
	}
}

func firstDate(dates []string) string {
	if len(dates) == 0 {
		return ""
	}
	return dates[0]
}
