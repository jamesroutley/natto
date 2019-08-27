package queue

import (
	"fmt"
	"sync"
)

type IDGenerator interface {
	ID() string
}

type IncrementingIDGenerator struct {
	next int64
	lock sync.Mutex
}

func (idgen *IncrementingIDGenerator) ID() string {
	idgen.lock.Lock()
	defer idgen.lock.Unlock()
	id := idgen.next
	idgen.next++
	return fmt.Sprint(id)
}
