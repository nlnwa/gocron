package gocron

import (
	"errors"
	"reflect"
	"runtime"
	"time"
)

type Job struct {

	// pause interval * unit bettween runs
	interval uint64

	// the job jobFunc to run, func[jobFunc]
	jobFunc string
	// time units, ,e.g. 'minutes', 'hours'...
	unit string
	// optional time at which this job runs
	atTime string

	// datetime of last run
	lastRun time.Time
	// datetime of next run
	nextRun time.Time
	// cache the period between last an next run
	period time.Duration

	// Specific day of the week to start on
	startDay time.Weekday

	// Map for the function task store
	funcs map[string]interface{}

	// Map for function and  params of function
	fparams map[string]([]interface{})
}

// Create a new job with the time interval.
func NewJob(intervel uint64) *Job {
	return &Job{
		intervel,
		"", "", "",
		time.Unix(0, 0),
		time.Unix(0, 0), 0,
		time.Sunday,
		make(map[string]interface{}),
		make(map[string]([]interface{})),
	}
}

// True if the job should be run now
func (j *Job) shouldRun() bool {
	return time.Now().After(j.nextRun)
}

//Run the job and immdiately reschedulei it
func (j *Job) run() (result []reflect.Value, err error) {
	f := reflect.ValueOf(j.funcs[j.jobFunc])
	params := j.fparams[j.jobFunc]
	if len(params) != f.Type().NumIn() {
		err = errors.New("The number of param is not adapted.")
		return
	}
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	result = f.Call(in)
	j.lastRun = time.Now()
	j.scheduleNextRun()
	return
}

// for given function fn , get the name of funciton.
func getFunctionName(fn interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf((fn)).Pointer()).Name()
}

// Specifies the jobFunc that should be called every time the job runs
//
func (j *Job) Do(jobFun interface{}, params ...interface{}) {
	typ := reflect.TypeOf(jobFun)
	if typ.Kind() != reflect.Func {
		panic("only function can be schedule into the job queue.")
	}

	fname := getFunctionName(jobFun)
	j.funcs[fname] = jobFun
	j.fparams[fname] = params
	j.jobFunc = fname
	//schedule the next run
	j.scheduleNextRun()
}

//	s.Every(1).Day().At("10:30").Do(task)
//	s.Every(1).Monday().At("10:30").Do(task)
func (j *Job) At(t string) *Job {
	hour := int((t[0]-'0')*10 + (t[1] - '0'))
	min := int((t[3]-'0')*10 + (t[4] - '0'))
	if hour < 0 || hour > 23 || min < 0 || min > 59 {
		panic("time format error.")
	}
	// time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	mock := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), int(hour), int(min), 0, 0, loc)

	if j.unit == "days" {
		if time.Now().After(mock) {
			j.lastRun = mock
		} else {
			j.lastRun = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-1, hour, min, 0, 0, loc)
		}
	} else if j.unit == "weeks" {
		if time.Now().After(mock) {
			i := mock.Weekday() - j.startDay
			if i < 0 {
				i = 7 + i
			}
			j.lastRun = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-int(i), hour, min, 0, 0, loc)
		} else {
			j.lastRun = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-7, hour, min, 0, 0, loc)
		}
	}
	return j
}

//Compute the instant when this job should run next
func (j *Job) scheduleNextRun() {
	if j.lastRun == time.Unix(0, 0) {
		if j.unit == "weeks" {
			i := time.Now().Weekday() - j.startDay
			if i < 0 {
				i = 7 + i
			}
			j.lastRun = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-int(i), 0, 0, 0, 0, loc)

		} else {
			j.lastRun = time.Now()
		}
	}

	if j.period != 0 {
		// translate all the units to the Seconds
		j.nextRun = j.lastRun.Add(j.period * time.Second)
	} else {
		switch j.unit {
		case "minutes":
			j.period = time.Duration(j.interval * 60)
			break
		case "hours":
			j.period = time.Duration(j.interval * 60 * 60)
			break
		case "days":
			j.period = time.Duration(j.interval * 60 * 60 * 24)
			break
		case "weeks":
			j.period = time.Duration(j.interval * 60 * 60 * 24 * 7)
			break
		case "seconds":
			j.period = time.Duration(j.interval)
		}
		j.nextRun = j.lastRun.Add(j.period * time.Second)
	}
}

// the follow functions set the job's unit with seconds,minutes,hours...

// Set the unit with second
func (j *Job) Second() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	job = j.Seconds()
	return
}

// Set the unit with seconds
func (j *Job) Seconds() (job *Job) {
	j.unit = "seconds"
	return j
}

// Set the unit  with minute, which interval is 1
func (j *Job) Minute() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	job = j.Minutes()
	return
}

//set the unit with minute
func (j *Job) Minutes() (job *Job) {
	j.unit = "minutes"
	return j
}

//set the unit with hour, which interval is 1
func (j *Job) Hour() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	job = j.Hours()
	return
}

// Set the unit with hours
func (j *Job) Hours() (job *Job) {
	j.unit = "hours"
	return j
}

// Set the job's unit with day, which interval is 1
func (j *Job) Day() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	job = j.Days()
	return
}

// Set the job's unit with days
func (j *Job) Days() *Job {
	j.unit = "days"
	return j
}

/*
// Set the unit with week, which the interval is 1
func (j *Job) Week() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	job = j.Weeks()
	return
}

*/

// s.Every(1).Monday().Do(task)
// Set the start day with Monday
func (j *Job) Monday() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	j.startDay = 1
	job = j.Weeks()
	return
}

// Set the start day with Tuesday
func (j *Job) Tuesday() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	j.startDay = 2
	job = j.Weeks()
	return
}

// Set the start day woth Wednesday
func (j *Job) Wednesday() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	j.startDay = 3
	job = j.Weeks()
	return
}

// Set the start day with thursday
func (j *Job) Thursday() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	j.startDay = 4
	job = j.Weeks()
	return
}

// Set the start day with friday
func (j *Job) Friday() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	j.startDay = 5
	job = j.Weeks()
	return
}

// Set the start day with saturday
func (j *Job) Saturday() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	j.startDay = 6
	job = j.Weeks()
	return
}

// Set the start day with sunday
func (j *Job) Sunday() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	j.startDay = 0
	job = j.Weeks()
	return
}

//Set the units as weeks
func (j *Job) Weeks() *Job {
	j.unit = "weeks"
	return j
}
