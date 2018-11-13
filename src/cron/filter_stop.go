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
	//f.row.lock.RLock()
	//defer f.row.lock.RUnlock()
	return 1 == atomic.LoadInt64(&f.row.Stop)// == 1
}
