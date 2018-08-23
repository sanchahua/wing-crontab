package cron

type StopFilter struct {
	row *CronEntity
}
func StopMiddleware() FilterMiddleWare {
	return func(entity *CronEntity) IFilter {
		return &StopFilter{entity}
	}
}

func (f *StopFilter) Stop() bool {
	f.row.lock.RLock()
	defer f.row.lock.RUnlock()
	return f.row.Stop
}
