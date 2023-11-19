package apitools

import (
	"fmt"
	"sort"
	"time"

	"github.com/nathan-osman/go-sunrise"
)

type Location struct {
	Latitude  float64
	Longitude float64
}

type SunPos int

const (
	Sunrise SunPos = iota
	Sunset
)

type scheduleSun struct {
	Location Location
	SunPos   SunPos
	Days     []time.Weekday
}

func ScheduleSunpos(location Location, sunriseSunset SunPos, days ...time.Weekday) scheduleSun {
	sort.Slice(days, func(i, j int) bool {
		return days[i] < days[j]
	})
	return scheduleSun{
		Location: location,
		SunPos:   sunriseSunset,
		Days:     days,
	}
}

func (s scheduleSun) seconds(t time.Time) int {
	t2 := s.OnDay(t)
	return t2.Hour()*3600 + t2.Minute()*60 + t2.Second()
}

func (s scheduleSun) String() string {
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
	sunpos := ""
	switch s.SunPos {
	case Sunrise:
		sunpos = "sunrise"
	case Sunset:
		sunpos = "sunset"
	}
	return fmt.Sprintf("%s%s", sunpos, days)
}

func (s scheduleSun) IsDay(t time.Time) bool {
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

func (s scheduleSun) OnDay(t time.Time) (t2 time.Time) {
	rise, set := sunrise.SunriseSunset(s.Location.Latitude, s.Location.Longitude, t.Year(), t.Month(), t.Day())
	switch s.SunPos {
	case Sunrise:
		t2 = rise
	case Sunset:
		t2 = set
	}
	return t2
}

func (s scheduleSun) Before(t time.Time) bool {
	st := s.OnDay(t)
	return st.Before(t)
}

func (s scheduleSun) After(t time.Time) bool {
	st := s.OnDay(t)
	return st.After(t) || st.Equal(t)
}

func (s scheduleSun) BeforeAndAfter(t time.Time) (before, after bool) {
	return s.Before(t), s.After(t)
}
