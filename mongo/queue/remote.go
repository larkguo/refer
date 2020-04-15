package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/mongodb/amboy"
	"github.com/mongodb/amboy/dependency"
	"github.com/mongodb/amboy/job"
	"github.com/mongodb/amboy/queue"
	"github.com/mongodb/amboy/registry"
)

var g_MyType string = "mytype"

func init() {
	registry.AddJobType(g_MyType, func() amboy.Job { return newSleepJob() })
	fmt.Println("Registry type:", g_MyType)
}

func main() {
	// context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// queue
	var err error
	creatOpts := queue.MongoDBQueueCreationOptions{}
	creatOpts.Name = "MongoDBQueue"
	creatOpts.Size = 1
	creatOpts.MDB = queue.DefaultMongoDBOptions()
	creatOpts.MDB.DB = "MongoQueueDB"
	creatOpts.MDB.Priority = true
	creatOpts.MDB.CheckWaitUntil = true
	creatOpts.MDB.URI = "mongodb://localhost:27017"
	queue, err := queue.NewMongoDBQueue(ctx, creatOpts)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("Start queue:", queue.ID())
	queue.Start(ctx)

	// producer
	producer := &JobProducer{
		q: queue,
	}
	producer.Start(ctx)

	// cancel
	interrupter := &Interrupter{cancel}
	interrupter.Watch()
}

type SleepJob struct {
	job.Base
	Sleep time.Duration
}

func newSleepJob() *SleepJob {
	j := &SleepJob{
		Sleep: 10 * time.Second,
		Base: job.Base{
			JobType: amboy.JobType{
				Name:    g_MyType,
				Version: 0,
			},
		},
	}
	j.SetDependency(dependency.NewAlways())
	j.SetID("myjob")
	return j
}
func (j *SleepJob) Run(ctx context.Context) {
	// customer
	defer j.MarkComplete()
	fmt.Println("Job callback run:", j.ID())
	if j.Sleep == 0 {
		return
	}
	timer := time.NewTimer(j.Sleep)
	defer timer.Stop()
	select {
	case <-timer.C:
		fmt.Println("Job after sleep:", j.Sleep)
		return
	case <-ctx.Done():
		fmt.Println("Job context done!")
		return
	}
}

type JobProducer struct {
	q amboy.Queue
}

func (p *JobProducer) Start(ctx context.Context) {
	go func() {
		j := newSleepJob()
		if err := p.q.Put(ctx, j); err != nil {
			fmt.Printf("error: %v\n", err)
		} else {
			fmt.Println("Job producer put:", j.ID())
		}
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Job producer context done!")
				return
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
		fmt.Println("Ctrl-C,stop JobProducer and SleepJob...")
		i.cancel() // Trigger JobProducer and SleepJob exit
		time.Sleep(5 * time.Second)
		break
	}
}

