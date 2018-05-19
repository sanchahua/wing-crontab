package crontab

type IFilter interface {
	Check() bool
}
