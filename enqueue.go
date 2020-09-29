package workers

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

const (
	// NanoSecondPrecision -
	NanoSecondPrecision = 1000000000.0
	// ScheduledJobsKey -
	ScheduledJobsKey = "schedule"
)

// EnqueueData enqueue data
type EnqueueData struct {
	Queue      string      `json:"queue,omitempty"`
	Class      string      `json:"class"`
	Args       interface{} `json:"args"`
	Jid        string      `json:"jid"`
	EnqueuedAt float64     `json:"enqueued_at"`
	EnqueueOptions
}

// EnqueueOptions enqueue options
type EnqueueOptions struct {
	RetryCount int     `json:"retry_count,omitempty"`
	Retry      bool    `json:"retry,omitempty"`
	At         float64 `json:"at,omitempty"`
}

func generateJid() string {
	// Return 12 random bytes as 24 character hex
	b := make([]byte, 12)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", b)
}

// Enqueue simple enqueue func
func Enqueue(keyword, queue, class string, args interface{}) (string, error) {
	return EnqueueWithOptions(keyword, queue, class, args, EnqueueOptions{At: nowToSecondsWithNanoPrecision()})
}

// EnqueueIn enqueue in
func EnqueueIn(keyword, queue, class string, in float64, args interface{}) (string, error) {
	return EnqueueWithOptions(keyword, queue, class, args, EnqueueOptions{At: nowToSecondsWithNanoPrecision() + in})
}

// EnqueueAt enqueue at
func EnqueueAt(keyword, queue, class string, at time.Time, args interface{}) (string, error) {
	return EnqueueWithOptions(keyword, queue, class, args, EnqueueOptions{At: timeToSecondsWithNanoPrecision(at)})
}

// EnqueueWithOptions enqueue with options
func EnqueueWithOptions(keyword, queue, class string, args interface{}, opts EnqueueOptions) (string, error) {
	now := nowToSecondsWithNanoPrecision()
	data := EnqueueData{
		Queue:          queue,
		Class:          class,
		Args:           args,
		Jid:            generateJid(),
		EnqueuedAt:     now,
		EnqueueOptions: opts,
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	if now < opts.At {
		err := enqueueAt(keyword, data.At, bytes)
		return data.Jid, err
	}

	pool, ok := Config.Pools[keyword]
	if !ok {
		return "", fmt.Errorf("are you sure? please ensure the conn keyword")
	}
	conn := pool.Pool.Get()
	defer conn.Close()

	_, err = conn.Do("sadd", pool.Namespace+"queues", queue)
	if err != nil {
		return "", err
	}
	queue = pool.Namespace + "queue:" + queue
	_, err = conn.Do("rpush", queue, bytes)
	if err != nil {
		return "", err
	}

	return data.Jid, nil
}

func enqueueAt(keyword string, at float64, bytes []byte) error {
	pool, ok := Config.Pools[keyword]
	if !ok {
		return fmt.Errorf("are you sure? please ensure the conn keyword")
	}
	conn := pool.Pool.Get()
	defer conn.Close()

	_, err := conn.Do(
		"zadd",
		pool.Namespace+ScheduledJobsKey, at, bytes,
	)
	if err != nil {
		return err
	}

	return nil
}

func timeToSecondsWithNanoPrecision(t time.Time) float64 {
	return float64(t.UnixNano()) / NanoSecondPrecision
}

func durationToSecondsWithNanoPrecision(d time.Duration) float64 {
	return float64(d.Nanoseconds()) / NanoSecondPrecision
}

func nowToSecondsWithNanoPrecision() float64 {
	return timeToSecondsWithNanoPrecision(time.Now())
}
