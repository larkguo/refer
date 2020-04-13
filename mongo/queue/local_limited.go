package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"golang.org/x/time/rate"

	"github.com/mongodb/amboy"
	"github.com/mongodb/amboy/job"
	"github.com/mongodb/amboy/queue"
)

func main() {
	queue := queue.NewLocalLimitedSize(1, 2)
	ctx, cancel := context.WithCancel(context.Background())
	queue.Start(ctx)
	
	// producer
	producer := &JobProducer{
		lim: rate.NewLimiter(1, 1),
		q:   queue,
	}
	producer.Start(ctx)
	
	interrupter := &Interrupter{cancel}
	interrupter.Watch()
}

type TimeJob struct {
	job.Base
}

func (job *TimeJob) ID() string {
	return time.Now().Format(time.RFC3339)
}
func (job *TimeJob) Run(ctx context.Context) {
        // customer 
	fmt.Printf("Time is %s\n", time.Now().Format(time.RFC3339))
	job.MarkComplete()
}

type JobProducer struct {
	lim *rate.Limiter
	q   amboy.Queue
}

func (p *JobProducer) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Closing job producer")
				return
			default:
				if ok := p.lim.Allow(); ok {
					if err := p.q.Put(ctx, &TimeJob{}); err != nil {
						fmt.Printf("error: %v\n", err)
					} else {
						fmt.Println("Put TimeJob")
					}
				}
			}
		}
	}()
}

type Interrupter struct {
	cancel context.CancelFunc
}

func (i *Interrupter) Watch() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for _ = range c {
		fmt.Println(" Ctrl-C ... ")
		i.cancel()
		break
	}
}
