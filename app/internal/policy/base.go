package policy

import (
	"time"
)

type Generator interface {
	GenerateID() string
}

type Clock interface {
	Now() time.Time
}

type BasePolicy struct {
	Generator
	Clock
}

func NewBasePolicy(generator Generator, clock Clock) *BasePolicy {
	return &BasePolicy{
		Generator: generator,
		Clock:     clock,
	}
}
