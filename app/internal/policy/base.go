package policy

import (
	"time"

	"github.com/google/uuid"
)

type Clock interface {
	Now() time.Time
}

type IDGenerator interface {
	Generate() string
}
type UUIDv7Generator struct {
	clock Clock
}

func NewUUIDv7Generator(clock Clock) *UUIDv7Generator {
	return &UUIDv7Generator{clock: clock}
}

func (g *UUIDv7Generator) Generate() string {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.New().String()
	}
	return id.String()
}

type BasePolicy struct {
	IDGenerator
	Clock
}

func NewBasePolicy(generator IDGenerator, clock Clock) *BasePolicy {
	return &BasePolicy{
		IDGenerator: generator,
		Clock:       clock,
	}
}
