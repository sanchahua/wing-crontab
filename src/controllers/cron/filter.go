package cron

type IFilter interface {
	Stop() bool
}
