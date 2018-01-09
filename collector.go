package gocron

import "sort"

type Collector interface {
	GetJobs() []*Job
}

// default in-memory job collector where
type defaultCollector struct {
	jobs []*Job
}

// Get jobs to be collected.
func (dc *defaultCollector) GetJobs() []*Job {
	return dc.jobs
}

// Create new default job collector.
func newDefaultCollector() Collector {
	return &defaultCollector{
		make([]*Job, 0),
	}
}

// Add appends job to list of jobs and sort list by time.
func (c *defaultCollector) add(job *Job) {
	c.jobs = append(c.jobs, job)
	sort.Sort(ByTime(c.jobs))
}
