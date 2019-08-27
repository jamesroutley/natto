package queue

import (
	"sync"

	"github.com/jamesroutley/natto/domain"
)

// Package queue implements a FIFO queue

type Message struct {
	ID       string
	Job      *domain.Job
	Attempts int
}

type Queue struct {
	// TODO: make into a RWMutex
	// Next  chan<- *Message
	lock       sync.Mutex
	head       *Message
	tail       *Message
	idgen      IDGenerator
	messages   chan *Message
	inProgress map[string]*Message
	// This channel is closed when the queue is empty of all messages
	// (including inProgress messages)
	empty chan bool
}

func NewQueue() *Queue {
	return &Queue{
		idgen:      &IncrementingIDGenerator{},
		messages:   make(chan *Message, 1000),
		inProgress: make(map[string]*Message),
		empty:      make(chan bool, 1),
	}
}

func (q *Queue) Add(job *domain.Job) {
	// TODO: return error if channel is full, rather than blocking
	message := q.NewMessage(job)
	q.messages <- message
}

// Next returns the next message that isn't currently being worked on
func (q *Queue) Next() *Message {
	message := <-q.messages

	// Lock after getting a message so we don't block other functions while we wait
	q.lock.Lock()
	defer q.lock.Unlock()

	// Add message to in-progress map
	q.inProgress[message.ID] = message

	return message
}

func (q *Queue) Delete(message *Message) {
	q.lock.Lock()
	defer q.lock.Unlock()
	delete(q.inProgress, message.ID)

	if len(q.messages) == 0 && len(q.inProgress) == 0 {
		q.empty <- true
	}
}

func (q *Queue) Error(message *Message) {
	q.lock.Lock()
	defer q.lock.Unlock()

	// Increment attempts, push it to the back of the queue
	message.Attempts++

	q.messages <- message
	delete(q.inProgress, message.ID)
}

// Wait blocks until the queue is empty of all messages
func (q *Queue) Wait() {
	<-q.empty
}

func (q *Queue) NewMessage(job *domain.Job) *Message {
	return &Message{
		ID:  q.idgen.ID(),
		Job: job,
	}
}
