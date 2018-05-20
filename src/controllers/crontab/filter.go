package crontab

type IFilter interface {
	Stop() bool
}
