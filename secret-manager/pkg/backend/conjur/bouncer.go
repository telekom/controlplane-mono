package conjur

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	registerOnce = sync.Once{}
	queueLength  = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "bouncer_runnable_queue_length",
			Help: "Number of runnables in the queue",
		},
		[]string{"queue"},
	)

	timeInQueue = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "bouncer_runnable_time_in_queue",
			Help:    "Time a runnable spent in the queue",
			Buckets: []float64{0.1, 0.5, 1, 2},
		},
		[]string{"queue", "status"},
	)

	ErrQueueFull = fmt.Errorf("queue is full")
)

const (
	defaultQueueSize        = 10
	defaultBlockingDuration = 2 * time.Second
)

func RegisterMetrics(reg prometheus.Registerer) {
	registerOnce.Do(func() {
		reg.MustRegister(queueLength)
		reg.MustRegister(timeInQueue)
	})
}

type contextKey string

const startKey contextKey = "start"

type Runnable func(ctx context.Context) error

type Bouncer interface {
	Run(ctx context.Context, runnable Runnable) <-chan error
	StartN(ctx context.Context, workerCount int)
}

type bouncer struct {
	Name                string
	RunnableQueue       chan Runnable
	MaxBlockingDuration time.Duration
}

func NewBouncer(queueSize int, blockingDur time.Duration) Bouncer {
	RegisterMetrics(prometheus.DefaultRegisterer)
	return &bouncer{
		Name:                "default",
		RunnableQueue:       make(chan Runnable, queueSize),
		MaxBlockingDuration: blockingDur,
	}
}

func NewDefaultBouncer() Bouncer {
	return NewBouncer(defaultQueueSize, defaultBlockingDuration)
}

func (b *bouncer) Run(ctx context.Context, runnable Runnable) <-chan error {
	log := logr.FromContextOrDiscard(ctx)

	done := make(chan error, 1)

	// timedCtx is used to set a timeout for enqueuing the runnable
	var timedCtx context.Context
	if b.MaxBlockingDuration > 0 {
		var cancel context.CancelFunc
		timedCtx, cancel = context.WithTimeout(ctx, b.MaxBlockingDuration)
		defer cancel()
	} else {
		timedCtx = ctx
	}

	ctx = context.WithValue(ctx, startKey, time.Now())

	wrappedRunnable := func(_ context.Context) error {
		start := ctx.Value(startKey).(time.Time)
		duration := time.Since(start).Seconds()
		timeInQueue.WithLabelValues(b.Name, "waiting").Observe(duration)
		err := runnable(ctx)
		if err != nil {
			log.V(1).Info("runnable failed", "error", err)
		}
		done <- err
		close(done)

		duration = time.Since(start).Seconds()
		status := "success"
		if err != nil {
			status = "error"
		}
		timeInQueue.WithLabelValues(b.Name, status).Observe(duration)
		return nil
	}
	select {
	case <-timedCtx.Done():
		done <- ErrQueueFull
		close(done)

	case b.RunnableQueue <- wrappedRunnable:
		log.V(1).Info("runnable enqueued")
		queueLength.WithLabelValues(b.Name).Inc()
	}
	return done
}

func (b *bouncer) StartN(ctx context.Context, n int) {
	for range n {
		b.startWorker(ctx)
	}
}

func (b *bouncer) startWorker(ctx context.Context) {
	go func(ctx context.Context) {
		log := logr.FromContextOrDiscard(ctx)
		defer func() {
			if r := recover(); r != nil {
				log.Error(fmt.Errorf("%v", r), "panic in bouncer worker")
			}
		}()
		for {
			select {
			case runnable := <-b.RunnableQueue:
				queueLength.WithLabelValues(b.Name).Dec()
				err := runnable(nil) // wrapped Runnable already has context
				if err != nil {
					log.Error(err, "failed to run runnable")
				}

			case <-ctx.Done():
				return
			}
		}
	}(ctx)
}
