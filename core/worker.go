package core

import "time"

type Worker interface {
	Schedule() string
	Ready(now time.Time) bool
	Execute()
}
