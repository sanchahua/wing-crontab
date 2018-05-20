package crontab

import (
	"time"
)
type TimeFilter struct {
	row *CronEntity
	next IFilter
}
func TimeMiddleware(next IFilter) CronEntityMiddleWare {
	return func(entity *CronEntity) IFilter {
		return &TimeFilter{row:entity, next:next}
	}
}

func (f *TimeFilter) Stop() bool {
	if f.next.Stop() {
		return true
	}

	if f.row.EndTime <= 0 {
		return false
	}

	current := time.Now().Unix()
	if current >= f.row.StartTime && current < f.row.EndTime {
		return false
	}
	return true
}

