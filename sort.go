package gocron

type ByTime []*Job

func (s ByTime) Len() int {
	return len(s)
}

func (s ByTime) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByTime) Less(i, j int) bool {
	if s[i].nextRun.IsZero() {
		return false
	}
	if s[j].nextRun.IsZero() {
		return true
	}
	return s[j].nextRun.After(s[i].nextRun)
}
