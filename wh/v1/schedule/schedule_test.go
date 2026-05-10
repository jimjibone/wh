package schedule_test

import (
	"testing"
	"time"

	"github.com/jimjibone/woodhouse-4/wh/v1/schedule"
	"github.com/nathan-osman/go-sunrise"
)

func TestScheduleTime(t *testing.T) {
	scheduleOn := schedule.ScheduleTime(6, 30, 00)
	scheduleOff := schedule.ScheduleTime(4, 55, 00)

	t1 := time.Date(2022, time.November, 10, 4, 30, 0, 0, time.Local)
	t2 := time.Date(2022, time.November, 10, 4, 55, 0, 0, time.Local)
	t3 := time.Date(2022, time.November, 10, 5, 30, 0, 0, time.Local)
	t4 := time.Date(2022, time.November, 10, 6, 30, 0, 0, time.Local)
	t5 := time.Date(2022, time.November, 10, 6, 31, 0, 0, time.Local)
	format := "15:04:05"

	before, after := scheduleOn.BeforeAndAfter(t1)
	t.Logf("scheduleOn: %s, t1: %s - before: %t, after: %t", scheduleOn, t1.Format(format), before, after)
	before, after = scheduleOn.BeforeAndAfter(t2)
	t.Logf("scheduleOn: %s, t2: %s - before: %t, after: %t", scheduleOn, t2.Format(format), before, after)
	before, after = scheduleOn.BeforeAndAfter(t3)
	t.Logf("scheduleOn: %s, t3: %s - before: %t, after: %t", scheduleOn, t3.Format(format), before, after)
	before, after = scheduleOn.BeforeAndAfter(t4)
	t.Logf("scheduleOn: %s, t4: %s - before: %t, after: %t", scheduleOn, t4.Format(format), before, after)
	before, after = scheduleOn.BeforeAndAfter(t5)
	t.Logf("scheduleOn: %s, t5: %s - before: %t, after: %t", scheduleOn, t5.Format(format), before, after)

	before, after = scheduleOff.BeforeAndAfter(t1)
	t.Logf("scheduleOff: %s, t1: %s - before: %t, after: %t", scheduleOff, t1.Format(format), before, after)
	before, after = scheduleOff.BeforeAndAfter(t2)
	t.Logf("scheduleOff: %s, t2: %s - before: %t, after: %t", scheduleOff, t2.Format(format), before, after)
	before, after = scheduleOff.BeforeAndAfter(t3)
	t.Logf("scheduleOff: %s, t3: %s - before: %t, after: %t", scheduleOff, t3.Format(format), before, after)
	before, after = scheduleOff.BeforeAndAfter(t4)
	t.Logf("scheduleOff: %s, t4: %s - before: %t, after: %t", scheduleOff, t4.Format(format), before, after)
	before, after = scheduleOff.BeforeAndAfter(t5)
	t.Logf("scheduleOff: %s, t5: %s - before: %t, after: %t", scheduleOff, t5.Format(format), before, after)
}

func TestScheduleTimeStr(t *testing.T) {
	_, err := schedule.ScheduleTimeStr("fred")
	if err == nil {
		t.Errorf(`ScheduleTimeStr should return error for "fred"`)
	}
	_, err = schedule.ScheduleTimeStr("one:two:three")
	if err == nil {
		t.Errorf(`ScheduleTimeStr should return error for "one:two:three"`)
	}
	sched, err := schedule.ScheduleTimeStr("10:52")
	if err != nil {
		t.Errorf("ScheduleTimeStr returned error for %q: %s", "10:52", err)
	} else if sched.Hour != 10 || sched.Minute != 52 || sched.Second != 00 {
		t.Errorf("ScheduleTimeStr returned incorrect time for %q: %s", "10:52", sched)
	}
	sched, err = schedule.ScheduleTimeStr("11:39:56")
	if err != nil {
		t.Errorf("ScheduleTimeStr returned error for %q: %s", "11:39:56", err)
	} else if sched.Hour != 11 || sched.Minute != 39 || sched.Second != 56 || len(sched.Days) != 0 {
		t.Errorf("ScheduleTimeStr returned incorrect time for %q: %s", "11:39:56", sched)
	}
	sched, err = schedule.ScheduleTimeStr("12:40:57", time.Monday)
	if err != nil {
		t.Errorf("ScheduleTimeStr returned error for %q: %s", "12:40:57 [mon]", err)
	} else if sched.Hour != 12 || sched.Minute != 40 || sched.Second != 57 || len(sched.Days) != 1 || sched.Days[0] != time.Monday {
		t.Errorf("ScheduleTimeStr returned incorrect time for %q: %s", "12:40:57 [mon]", sched)
	}
}

func TestSchedule(t *testing.T) {
	sched := schedule.Schedule[float64]{
		{Time: schedule.MustScheduleTimeStr("06:30:00", time.Monday), Value: 18.0},
		{Time: schedule.MustScheduleTimeStr("07:30:00", schedule.Weekdays...), Value: 14.0},
		{Time: schedule.MustScheduleTimeStr("16:30:00", time.Monday, time.Tuesday), Value: 28.0},
		{Time: schedule.MustScheduleTimeStr("22:00:00", time.Wednesday), Value: 14.0},
	}

	now := time.Date(2023, 01, 31, 17, 10, 36, 0, time.Local) // tuesday

	currentSched, currentTime := sched.GetCurrent(now)
	t.Logf("current sched: %s, time: %s", currentSched, currentTime)

	nextSched, nextTime := sched.GetNext(now)
	t.Logf("next sched: %s, time: %s", nextSched, nextTime)
}

func TestScheduleFirst(t *testing.T) {
	location := schedule.Location{
		Latitude:  51.47,
		Longitude: 0.00,
	}

	// now := time.Date(2023, 01, 31, 17, 10, 36, 0, time.UTC) // tuesday
	now := time.Date(2026, 04, 14, 17, 10, 36, 0, time.Local) // monday
	t.Log("now:", now)

	a := schedule.MustScheduleTimeStr("19:00:00")
	b := schedule.ScheduleSunpos(location, schedule.Sunset, -time.Hour)
	t.Logf("a: %s", a.OnDay(now))
	t.Logf("b: %s", b.OnDay(now))

	sched := schedule.ScheduleFirst(a, b)

	rise, set := sunrise.SunriseSunset(location.Latitude, location.Longitude, now.Year(), now.Month(), now.Day())
	t.Log("sunrise:", rise, "sunset:", set)

	res := sched.OnDay(now)
	if res.Hour() != 19 || res.Minute() != 0 || res.Second() != 0 {
		t.Errorf("ScheduleFirst returned incorrect time: %s", res)
	}
	t.Logf("first: %s", res)
}
