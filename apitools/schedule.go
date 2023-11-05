package apitools

import (
	"fmt"
	"sort"
	"time"
)

type Schedule[T any] []ScheduleEntry[T]

type ScheduleEntry[T any] struct {
	Value T
	Time  scheduleTime
}

func (s ScheduleEntry[T]) String() string {
	return fmt.Sprintf("{%s: %v}", s.Time, s.Value)
}

func (s Schedule[T]) Sort() {
	sort.Slice(s, func(i, j int) bool {
		return s[i].Time.seconds() < s[j].Time.seconds()
	})
}

func (s Schedule[T]) GetCurrent(t time.Time) (ScheduleEntry[T], time.Time) {
	s.Sort()
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
	s.Sort()
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
