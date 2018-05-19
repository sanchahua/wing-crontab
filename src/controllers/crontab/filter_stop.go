package crontab

type StopFilter struct {
	row *CronEntity
}
func StopMiddleware() CronEntityMiddleWare {
	return func(entity *CronEntity) IFilter {
		return &StopFilter{entity}
	}
}

func (f *StopFilter) Check() bool {
	return f.row.Stop
}
