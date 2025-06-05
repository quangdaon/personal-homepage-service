package core

type Worker interface {
	Schedule() string
	Ready() bool
	Execute()
}
