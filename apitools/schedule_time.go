package apitools

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type scheduleTime struct {
	Hour   int
	Minute int
	Second int
}

func ScheduleTime(hour, min, sec int) scheduleTime {
	return scheduleTime{
		Hour:   hour,
		Minute: min,
		Second: sec,
	}
}

func ScheduleTimeStr(t string) (scheduleTime, error) {
	s := scheduleTime{}
	parts := strings.Split(t, ":")
	if len(parts) >= 2 {
		v, err := strconv.Atoi(parts[0])
		if err != nil {
			return s, fmt.Errorf("failed to parse first time component: %s", err)
		}
		s.Hour = v
		v, err = strconv.Atoi(parts[1])
		if err != nil {
			return s, fmt.Errorf("failed to parse second time component: %s", err)
		}
		s.Minute = v
		if len(parts) == 3 {
			v, err := strconv.Atoi(parts[2])
			if err != nil {
				return s, fmt.Errorf("failed to parse third time component: %s", err)
			}
			s.Second = v
		}
	} else {
		return s, fmt.Errorf("not enough time components")
	}
	return s, nil
}

func MustScheduleTimeStr(t string) scheduleTime {
	s, err := ScheduleTimeStr(t)
	if err != nil {
		panic(err)
	}
	return s
}

func (s scheduleTime) seconds() int {
	return s.Hour*3600 + s.Minute*60 + s.Second
}

func (s scheduleTime) String() string {
	return fmt.Sprintf("%02d:%02d:%02d", s.Hour, s.Minute, s.Second)
}

func (s scheduleTime) OnDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), s.Hour, s.Minute, s.Second, 0, t.Location())
}

func (s scheduleTime) Before(t time.Time) bool {
	st := s.OnDay(t)
	return st.Before(t)
}

func (s scheduleTime) After(t time.Time) bool {
	st := s.OnDay(t)
	return st.After(t) || st.Equal(t)
}

func (s scheduleTime) BeforeAndAfter(t time.Time) (before, after bool) {
	return s.Before(t), s.After(t)
}
