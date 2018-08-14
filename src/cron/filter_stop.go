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
	return f.row.Stop
}
