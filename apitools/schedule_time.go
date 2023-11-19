package apitools

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

type scheduleTime struct {
	Days   []time.Weekday
	Hour   int
	Minute int
	Second int
}

var Weekdays = []time.Weekday{
	time.Monday,
	time.Tuesday,
	time.Wednesday,
	time.Thursday,
	time.Friday,
}

var Weekend = []time.Weekday{
	time.Sunday,
	time.Saturday,
}

func ScheduleTime(hour, min, sec int, days ...time.Weekday) scheduleTime {
	sort.Slice(days, func(i, j int) bool {
		return days[i] < days[j]
	})
	return scheduleTime{
		Days:   days,
		Hour:   hour,
		Minute: min,
		Second: sec,
	}
}

func ScheduleTimeStr(t string, days ...time.Weekday) (scheduleTime, error) {
	sort.Slice(days, func(i, j int) bool {
		return days[i] < days[j]
	})
	s := scheduleTime{
		Days: days,
	}
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

func MustScheduleTimeStr(t string, days ...time.Weekday) scheduleTime {
	s, err := ScheduleTimeStr(t, days...)
	if err != nil {
		panic(err)
	}
	return s
}

func (s scheduleTime) seconds(t time.Time) int {
	return s.Hour*3600 + s.Minute*60 + s.Second
}

func (s scheduleTime) String() string {
	days := ""
	if len(s.Days) > 0 {
		weekdaysOnly := true
		weekendOnly := true
		days += " ["
		for i, d := range s.Days {
			if i > 0 {
				days += ","
			}
			switch d {
			case time.Sunday:
				days += "sun"
				weekdaysOnly = false
			case time.Monday:
				days += "mon"
				weekendOnly = false
			case time.Tuesday:
				days += "tue"
				weekendOnly = false
			case time.Wednesday:
				days += "wed"
				weekendOnly = false
			case time.Thursday:
				days += "thu"
				weekendOnly = false
			case time.Friday:
				days += "fri"
				weekendOnly = false
			case time.Saturday:
				days += "sat"
				weekdaysOnly = false
			}
		}
		if weekdaysOnly && len(s.Days) == 5 {
			days = " [weekdays]"
		} else if weekendOnly && len(s.Days) == 2 {
			days = " s[weekend]"
		}
		days += "]"
	}
	return fmt.Sprintf("%02d:%02d:%02d%s", s.Hour, s.Minute, s.Second, days)
}

func (s scheduleTime) IsDay(t time.Time) bool {
	if len(s.Days) == 0 {
		return true
	}
	for _, d := range s.Days {
		if t.Weekday() == d {
			return true
		}
	}
	return false
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
