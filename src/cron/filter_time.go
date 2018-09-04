package cron

import (
	"time"
	time2 "library/time"
)
type TimeFilter struct {
	row *CronEntity
	next IFilter
}
func TimeMiddleware(next IFilter) FilterMiddleWare {
	return func(entity *CronEntity) IFilter {
		return &TimeFilter{row:entity, next:next}
	}
}

func (f *TimeFilter) Stop() bool {
	if f.next.Stop() {
		return true
	}

	f.row.lock.RLock()
	defer f.row.lock.RUnlock()

	et := time2.StrToTime(f.row.EndTime)
	if et <= 0 {
		return false
	}

	st := time2.StrToTime(f.row.StartTime)
	current := time.Now().Unix()
	if current >= st && current < et {
		return false
	}
	return true
}

