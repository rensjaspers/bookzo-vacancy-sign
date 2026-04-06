package vacancy

import (
	"sync"
	"time"
)

type Store struct {
	mu          sync.RWMutex
	settings    Settings
	load        LoadResult
	hasLoad     bool
	lastSuccess time.Time
}

type snapshotState struct {
	settings    Settings
	load        LoadResult
	hasLoad     bool
	lastSuccess time.Time
}

func NewStore(settings Settings) *Store {
	return &Store{settings: settings}
}

func (s *Store) UpdateLoad(result LoadResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.load = result
	s.hasLoad = true
	s.updateLastSuccess(result)
}

func (s *Store) updateLastSuccess(result LoadResult) {
	if result.FetchFailed {
		return
	}
	s.lastSuccess = result.FetchedAt
}

func (s *Store) Snapshot(now time.Time) ViewModel {
	s.mu.RLock()
	state := s.snapshotState()
	s.mu.RUnlock()
	return buildViewModel(state, now)
}

func (s *Store) snapshotState() snapshotState {
	return snapshotState{
		settings:    s.settings,
		load:        s.load,
		hasLoad:     s.hasLoad,
		lastSuccess: s.lastSuccess,
	}
}

func buildViewModel(state snapshotState, now time.Time) ViewModel {
	index := PhraseIndex(now, state.settings.PhraseRotate, len(state.settings.Languages))
	language := LanguageAt(state.settings.Languages, index)
	available := effectiveAvailable(state)
	today := isAvailableToday(state, now)
	return ViewModel{
		HotelName:          state.settings.HotelName,
		Headline:           HeadlineFor(language, available, today),
		Subline:            availableSubline(state, language, now),
		Dark:               effectiveDark(state.settings, now),
		Available:          available,
		ShowSubline:        showSubline(state, now),
		ShowErrorHint:      showErrorHint(state),
		HeadlineScale:      state.settings.HeadlineScale,
		HeadlineCandidates: headlineCandidates(state.settings, available, today),
		SublineCandidates:  sublineCandidates(state, now),
	}
}

func headlineCandidates(settings Settings, available, today bool) []string {
	langs := settings.Languages
	if len(langs) == 0 {
		langs = []Language{defaultLanguage()}
	}
	out := make([]string, 0, len(langs))
	for _, lang := range langs {
		out = append(out, HeadlineFor(lang, available, today))
	}
	return out
}

func sublineCandidates(state snapshotState, now time.Time) []string {
	if !showSubline(state, now) {
		return nil
	}
	langs := state.settings.Languages
	if len(langs) == 0 {
		langs = []Language{defaultLanguage()}
	}
	first := state.load.FirstAvailableDate
	out := make([]string, 0, len(langs))
	for _, lang := range langs {
		out = append(out, SublineFor(lang, first))
	}
	return out
}

func effectiveAvailable(state snapshotState) bool {
	switch state.settings.OverrideMode {
	case OverrideAvailable:
		return true
	case OverrideUnavailable:
		return false
	default:
		return autoAvailable(state)
	}
}

func autoAvailable(state snapshotState) bool {
	if !state.hasLoad {
		return true
	}
	return state.load.AnyRoomAvailable
}

func isAvailableToday(state snapshotState, now time.Time) bool {
	if !effectiveAvailable(state) {
		return false
	}
	return todayMatch(state, now)
}

func todayMatch(state snapshotState, now time.Time) bool {
	today, _ := EffectiveToday(now, state.settings.TodayDateOverride)
	if state.settings.OverrideMode != OverrideAuto {
		return true
	}
	return firstDateMatches(state, today)
}

func firstDateMatches(state snapshotState, today time.Time) bool {
	return state.load.FirstAvailableDate == FormatYMD(today) && !state.load.FetchFailed
}

func availableSubline(state snapshotState, language Language, now time.Time) string {
	if !showSubline(state, now) {
		return ""
	}
	return SublineFor(language, state.load.FirstAvailableDate)
}

func showSubline(state snapshotState, now time.Time) bool {
	if !effectiveAvailable(state) || state.settings.OverrideMode != OverrideAuto {
		return false
	}
	return firstAvailableAfterRef(state, now)
}

func firstAvailableAfterRef(state snapshotState, now time.Time) bool {
	today, _ := EffectiveToday(now, state.settings.TodayDateOverride)
	refYMD := FormatYMD(today)
	first := state.load.FirstAvailableDate
	return first != "" && first > refYMD
}

func showErrorHint(state snapshotState) bool {
	if !state.settings.ShowErrorHint {
		return false
	}
	return state.settings.OverrideMode == OverrideAuto && state.load.FetchFailed
}

func effectiveDark(settings Settings, now time.Time) bool {
	switch settings.ThemeMode {
	case ThemeLight:
		return false
	case ThemeTime:
		return !insideLightWindow(settings.LightWindow, now)
	default:
		return true
	}
}

func insideLightWindow(window ThemeWindow, now time.Time) bool {
	minute := now.Hour()*60 + now.Minute()
	low, high := orderedWindow(window)
	return minute >= low && minute <= high
}

func orderedWindow(window ThemeWindow) (int, int) {
	if window.StartMinute < window.EndMinute {
		return window.StartMinute, window.EndMinute
	}
	return window.EndMinute, window.StartMinute
}
