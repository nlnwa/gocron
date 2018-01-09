package gocron

import (
	"errors"
	"github.com/nlnwa/kronasje"
	"reflect"
	"time"
)

type Schedule interface {
	Next(time.Time) time.Time
}

type Job struct {
	// schedule based on cron expression determines runtime
	Schedule Schedule

	// the tasks this job executes
	tasks []reflect.Value

	// the parameters that will be passed tasks upon execution
	tasksParams [][]reflect.Value

	// time of last run
	lastRun time.Time

	// time of next run
	nextRun time.Time
}

// Run job by calling the jobs task functions.
func (j *Job) run() {
	for i := range j.tasks {
		j.tasks[i].Call(j.tasksParams[i])
	}
}

// NewCronJob creates a new job based on a cron expression.
func NewCronJob(expression string) (*Job, error) {
	schedule, err := kronasje.Parse(expression)
	if err != nil {
		return nil, err
	}
	return NewJob(schedule.(Schedule)), nil
}

// NewJob creates a new job based on a schedule.
func NewJob(schedule Schedule) *Job {
	return &Job{
		Schedule: schedule,
	}
}

// Add a task function including it's parameters to the job.
func (j *Job) AddTask(task interface{}, params ...interface{}) *Job {
	// reflect the task and params to values
	taskValue := reflect.ValueOf(task)
	paramValues := make([]reflect.Value, len(params))
	for i := range params {
		paramValues[i] = reflect.ValueOf(params[i])
	}

	if taskValue.Type().NumIn() != len(paramValues) {
		panic(errors.New("mismatched number of task parameters"))
	} else if taskValue.Kind() != reflect.Func {
		panic(errors.New("task is not a function"))
	}

	// add the task and its params to the job
	j.tasks = append(j.tasks, taskValue)
	j.tasksParams = append(j.tasksParams, paramValues)

	return j
}
