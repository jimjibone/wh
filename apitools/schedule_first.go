package apitools

import (
	"fmt"
	"time"
)

type scheduleFirst struct {
	A schedulePoint
	B schedulePoint
}

func ScheduleFirst(a, b schedulePoint) scheduleFirst {
	return scheduleFirst{
		A: a,
		B: b,
	}
}

func (s scheduleFirst) which(t time.Time) schedulePoint {
	a := s.A.seconds(t)
	b := s.B.seconds(t)
	if a < b {
		return s.A
	}
	return s.B
}

func (s scheduleFirst) seconds(t time.Time) int {
	a := s.A.seconds(t)
	b := s.B.seconds(t)
	if a < b {
		return a
	}
	return b
}

func (s scheduleFirst) String() string {
	return fmt.Sprintf("A: %s, B: %s", s.A, s.B)
}

func (s scheduleFirst) IsDay(t time.Time) bool {
	return s.which(t).IsDay(t)
}

func (s scheduleFirst) OnDay(t time.Time) time.Time {
	return s.which(t).OnDay(t)
}

func (s scheduleFirst) Before(t time.Time) bool {
	return s.which(t).Before(t)
}

func (s scheduleFirst) After(t time.Time) bool {
	return s.which(t).After(t)
}

func (s scheduleFirst) BeforeAndAfter(t time.Time) (before, after bool) {
	first := s.which(t)
	return first.Before(t), first.After(t)
}
