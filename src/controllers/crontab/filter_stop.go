package crontab

type StopFilter struct {
	row *CronEntity
}
func StopMiddleware() CronEntityMiddleWare {
	return func(entity *CronEntity) IFilter {
		return &StopFilter{entity}
	}
}

func (f *StopFilter) Stop() bool {
	return f.row.Stop
}
