package queue

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/jamesroutley/natto/domain"
	"github.com/stretchr/testify/assert"
)

func TestQueue(t *testing.T) {
	numMessages := 5
	q := NewQueue()

	// Add items to the queue
	for i := 0; i < numMessages; i++ {
		q.Add(&domain.Job{URL: exampleURL(fmt.Sprint(i))})
	}

	// Read items from the queue, and delete them
	for i := 0; i < numMessages; i++ {
		message := q.Next()
		assert.Equal(t, message.Job.URL, exampleURL(fmt.Sprint(i)))
		q.Delete(message)
	}

	assert.Empty(t, len(q.messages))
	assert.Empty(t, len(q.inProgress))

	q.Wait()
}

func TestQueueError(t *testing.T) {
	q := NewQueue()
	job := &domain.Job{URL: exampleURL("xyz")}

	q.Add(job)

	message := q.Next()

	assert.Equal(t, message.Job, job)
	assert.Empty(t, len(q.messages))
	assert.Equal(t, len(q.inProgress), 1)

	q.Error(message)

	assert.Equal(t, len(q.messages), 1)
	assert.Empty(t, len(q.inProgress))

	message = q.Next()

	assert.Equal(t, message.Job, job)
	assert.Empty(t, len(q.messages))
	assert.Equal(t, len(q.inProgress), 1)

	q.Delete(message)

	assert.Empty(t, len(q.messages))
	assert.Empty(t, len(q.inProgress))

	q.Wait()
}

func exampleURL(path string) *url.URL {
	return &url.URL{
		Scheme: "https",
		Host:   "example.com",
		Path:   path,
	}
}
