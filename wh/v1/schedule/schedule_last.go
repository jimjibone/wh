package schedule

import (
	"fmt"
	"time"
)

type scheduleLast struct {
	A schedulePoint
	B schedulePoint
}

func ScheduleLast(a, b schedulePoint) scheduleLast {
	return scheduleLast{
		A: a,
		B: b,
	}
}

func (s scheduleLast) which(t time.Time) schedulePoint {
	a := s.A.seconds(t)
	b := s.B.seconds(t)
	if a > b {
		return s.A
	}
	return s.B
}

func (s scheduleLast) seconds(t time.Time) int {
	a := s.A.seconds(t)
	b := s.B.seconds(t)
	if a > b {
		return a
	}
	return b
}

func (s scheduleLast) String() string {
	return fmt.Sprintf("A: %s, B: %s", s.A, s.B)
}

func (s scheduleLast) IsDay(t time.Time) bool {
	return s.which(t).IsDay(t)
}

func (s scheduleLast) OnDay(t time.Time) time.Time {
	return s.which(t).OnDay(t)
}

func (s scheduleLast) Before(t time.Time) bool {
	return s.which(t).Before(t)
}

func (s scheduleLast) After(t time.Time) bool {
	return s.which(t).After(t)
}

func (s scheduleLast) BeforeAndAfter(t time.Time) (before, after bool) {
	last := s.which(t)
	return last.Before(t), last.After(t)
}
