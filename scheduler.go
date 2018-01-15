package gocron

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// Scheduler contains jobs and a loop to run the jobs
type Scheduler struct {
	sync.RWMutex
	location   *time.Location
	interval   time.Duration
	collectors []Collector
	isRunning  bool
	done       chan struct{}
}

// NewScheduler create a new scheduler.
func NewScheduler() *Scheduler {
	return &Scheduler{
		collectors: []Collector{newDefaultCollector()},
		location:   time.Local,
		interval:   time.Minute,
	}
}

func (s *Scheduler) AddJob(job *Job) {
	s.Lock()
	defer s.Unlock()
	s.collectors[0].(*defaultCollector).add(job)
}

func (s *Scheduler) AddCollector(collector Collector) *Scheduler {
	s.Lock()
	defer s.Unlock()
	s.collectors = append(s.collectors, collector)
	return s
}

// Interval sets the interval between the check for pending jobs.
func (s *Scheduler) Interval(d time.Duration) *Scheduler {
	s.Lock()
	defer s.Unlock()
	s.interval = d
	return s
}

// Location sets the location of the scheduler.
func (s *Scheduler) Location(location *time.Location) *Scheduler {
	s.Lock()
	defer s.Unlock()
	if s.isRunning {
		s.location = location
	}
	return s
}

// runPending runs all of the jobs pending now.
func (s *Scheduler) runPending(now time.Time) {
	s.Lock()
	defer s.Unlock()

	for i := range s.collectors {
		jobs := s.collectors[i].GetJobs()
		for _, job := range jobs {
			if job.nextRun.IsZero() {
				job.nextRun = job.Schedule.Next(now)
			}
			if job.nextRun.After(now) {
				continue
			}
			log.WithField("now", now).Infoln("Starting job")
			go job.run()
			job.lastRun = job.nextRun
			job.nextRun = job.Schedule.Next(now)
		}
	}
}

// Start scheduler.
func (s *Scheduler) Start() {
	s.Lock()
	defer s.Unlock()

	// only start the scheduler if it hasn't been started yet
	if s.isRunning {
		return
	} else {
		s.done = make(chan struct{})
		s.isRunning = true
	}

	go func() {
		ticker := time.NewTicker(s.interval)
		for {
			select {
			case <-s.done:
				ticker.Stop()
				s.done <- struct{}{}
				return
			case now := <-ticker.C:
				s.runPending(now.In(s.location))
			}
		}
	}()
}

// Stop scheduler.
func (s *Scheduler) Stop() {
	s.Lock()
	defer s.Unlock()
	if s.isRunning {
		s.done <- struct{}{}
		<-s.done
		s.isRunning = false
	}
}
