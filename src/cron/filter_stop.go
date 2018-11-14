package cron

import "sync/atomic"

type StopFilter struct {
	row *CronEntity
}
func StopMiddleware() FilterMiddleWare {
	return func(entity *CronEntity) IFilter {
		return &StopFilter{entity}
	}
}

func (f *StopFilter) Stop() bool {
	return 1 == atomic.LoadInt64(&f.row.Stop)
}
