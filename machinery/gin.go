/*
docker run -d -p 5672:5672 -p 15672:15672 --name rabbitmq rabbitmq:management
docker run --name mongo -p 27017:27017 -d --restart=always -v ~/mongo/db:/data/db -v /etc/localtime:/etc/localtime:ro --log-opt max-size=50m mongo
*/

package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/RichardKnop/machinery/v2"
	dbbackend "github.com/RichardKnop/machinery/v2/backends/mongo"
	amqpbroker "github.com/RichardKnop/machinery/v2/brokers/amqp"
	"github.com/RichardKnop/machinery/v2/config"
	eagerlock "github.com/RichardKnop/machinery/v2/locks/eager"
	"github.com/RichardKnop/machinery/v2/log"
	"github.com/RichardKnop/machinery/v2/tasks"
	"github.com/gin-gonic/gin"
)

func main() {
	go worker()

	r := gin.Default()
	r.GET("/concat", func(c *gin.Context) {
		ProcessConcat(c)
	})
	r.Run(":8000")
}

func Concat(p1, p2 string) (string, error) {
	var res string
	res += p1
	res += p2
	return res, nil
}

func startServer() (*machinery.Server, error) {
	cnf := &config.Config{
		Broker:          "amqp://guest:guest@localhost:5672/",
		DefaultQueue:    "task_queue",
		ResultBackend:   "mongodb://localhost:27017",
		ResultsExpireIn: 3600,
		AMQP: &config.AMQPConfig{
			Exchange:      "task_exchange",
			ExchangeType:  "direct",
			BindingKey:    "task_queue",
			PrefetchCount: 3,
		},
	}

	// Create server instance
	broker := amqpbroker.New(cnf)
	backend, err := dbbackend.New(cnf)
	if err != nil {
		log.ERROR.Printf("dbbackend.New() error %v\n", err.Error())
		return nil, err
	}
	lock := eagerlock.New()
	server := machinery.NewServer(cnf, broker, backend, lock)

	// Register tasks
	tasksMap := map[string]interface{}{
		"concat": Concat,
	}
	return server, server.RegisterTasks(tasksMap)
}

func worker() error {
	consumerTag := "machinery_worker"
	server, err := startServer()
	if err != nil {
		return err
	}
	worker := server.NewWorker(consumerTag, 0)
	errorHandler := func(err error) {
		log.ERROR.Println("I am an error handler:", err)
	}
	preTaskHandler := func(signature *tasks.Signature) {
		log.INFO.Println("I am a start of task handler for:", signature.Name)
	}
	postTaskHandler := func(signature *tasks.Signature) {
		log.INFO.Println("I am an end of task handler for:", signature.Name)
	}
	worker.SetPostTaskHandler(postTaskHandler)
	worker.SetErrorHandler(errorHandler)
	worker.SetPreTaskHandler(preTaskHandler)
	return worker.Launch()
}

func ProcessConcat(c *gin.Context) error {
	server, err := startServer()
	if err != nil {
		return err
	}
	var concatTask tasks.Signature
	var initTasks = func() {
		concatTask = tasks.Signature{
			Name: "concat",
			Args: []tasks.Arg{
				{Type: "string", Value: c.Query("p1")},
				{Type: "string", Value: c.Query("p2")},
			},
		}
	}
	initTasks()
	asyncResult, err := server.SendTask(&concatTask)
	if err != nil {
		fmt.Errorf("Could not send task: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return err
	}
	results, err := asyncResult.Get(time.Millisecond * 5)
	if err != nil {
		fmt.Errorf("Getting task result failed with error: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return err
	}
	log.INFO.Printf("result = %v\n", tasks.HumanReadableResults(results))

	c.String(http.StatusOK, tasks.HumanReadableResults(results))

	return nil
}
