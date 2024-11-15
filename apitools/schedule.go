package apitools

import (
	"context"
	"fmt"
	"sort"
	"time"
)

type Schedule[T any] []ScheduleEntry[T]

type ScheduleEntry[T any] struct {
	Value T
	Time  schedulePoint
}

type schedulePoint interface {
	seconds(t time.Time) int
	String() string
	IsDay(t time.Time) bool
	OnDay(t time.Time) time.Time
	Before(t time.Time) bool
	After(t time.Time) bool
	BeforeAndAfter(t time.Time) (before, after bool)
}

func (s ScheduleEntry[T]) String() string {
	return fmt.Sprintf("{%s: %v}", s.Time, s.Value)
}

func (s Schedule[T]) Sort(t time.Time) {
	sort.Slice(s, func(i, j int) bool {
		return s[i].Time.seconds(t) < s[j].Time.seconds(t)
	})
}

func (s Schedule[T]) GetCurrent(t time.Time) (ScheduleEntry[T], time.Time) {
	s.Sort(t)
	found := false
	var curr ScheduleEntry[T]
	for _, e := range s {
		if e.Time.IsDay(t) && e.Time.Before(t) {
			found = true
			curr = e
		}
	}
	for !found {
		// Try the previous day.
		t = time.Date(t.Year(), t.Month(), t.Day()-1, 23, 59, 59, 999999, t.Location())
		s.Sort(t)
		for _, e := range s {
			if e.Time.IsDay(t) && e.Time.Before(t) {
				found = true
				curr = e
			}
		}
	}
	return curr, curr.Time.OnDay(t)
}

func (s Schedule[T]) GetNext(t time.Time) (ScheduleEntry[T], time.Time) {
	s.Sort(t)
	found := false
	var curr ScheduleEntry[T]
	for _, e := range s {
		if e.Time.IsDay(t) && !e.Time.Before(t) {
			found = true
			curr = e
			break
		}
	}
	for !found {
		// Try the next day.
		t = time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, t.Location())
		s.Sort(t)
		for _, e := range s {
			if e.Time.IsDay(t) && !e.Time.Before(t) {
				found = true
				curr = e
				break
			}
		}
	}
	return curr, curr.Time.OnDay(t)
}

func (s Schedule[T]) Run(ctx context.Context, handler func(startTime time.Time, setting T)) {
	setting, startTime := s.GetCurrent(time.Now())
	handler(startTime, setting.Value)

	startTime = time.Now()
	for {
		setting, nextTime := s.GetNext(startTime)
		timer := time.NewTimer(nextTime.Sub(startTime))

		select {
		case <-ctx.Done():
			timer.Stop()
			return

		case now := <-timer.C:
			handler(now, setting.Value)
			startTime = now
		}
	}
}
