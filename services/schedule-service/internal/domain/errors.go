package domain

import "errors"

var (
	ErrScheduleNotFound      = errors.New("schedule not found")
	ErrScheduleAccessDenied  = errors.New("access denied to schedule")
	ErrInvalidSchedule       = errors.New("invalid schedule")
	ErrMissingCronExpression = errors.New("cron expression required for cron schedule type")
	ErrMissingRunAt          = errors.New("run_at required for one_shot schedule type")
	ErrInvalidAction         = errors.New("invalid action target")
	ErrPublishFailed         = errors.New("failed to publish schedule event")
)
