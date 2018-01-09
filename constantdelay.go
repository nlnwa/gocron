package gocron

import (
	"time"
)

type recurringSchedule struct {
	interval time.Duration
}

func Every(interval time.Duration) Schedule {
	return &recurringSchedule{
		interval: interval,
	}
}

func (r *recurringSchedule) Next(now time.Time) time.Time {
	return now.Add(r.interval)
}
