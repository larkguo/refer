/*
docker run -d -p 5672:5672 -p 15672:15672 --name rabbitmq rabbitmq:management
docker run --name mongo -p 27017:27017 -d --restart=always -v ~/mongo/db:/data/db -v /etc/localtime:/etc/localtime:ro --log-opt max-size=50m mongo
*/

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/RichardKnop/machinery/v2"
	dbbackend "github.com/RichardKnop/machinery/v2/backends/mongo"
	amqpbroker "github.com/RichardKnop/machinery/v2/brokers/amqp"
	"github.com/RichardKnop/machinery/v2/config"
	eagerlock "github.com/RichardKnop/machinery/v2/locks/eager"
	"github.com/RichardKnop/machinery/v2/log"
	"github.com/RichardKnop/machinery/v2/tasks"
	"github.com/urfave/cli"
)

var app *cli.App

func init() {
	app = cli.NewApp()
	app.Name = "machinery"
	app.Usage = "machinery worker and send example tasks with machinery send"
	app.Version = "0.0.1"
}

func main() {
	app.Commands = []cli.Command{ // Set the CLI app commands
		{
			Name:  "worker",
			Usage: "launch machinery worker",
			Action: func(c *cli.Context) error {
				if err := worker(); err != nil {
					return cli.NewExitError(err.Error(), 1)
				}
				return nil
			},
		}, {
			Name:  "send",
			Usage: "send example tasks ",
			Action: func(c *cli.Context) error {
				if err := send(); err != nil {
					return cli.NewExitError(err.Error(), 1)
				}
				return nil
			},
		},
	}
	app.Run(os.Args) // Run the CLI app
}

func Concat(strs []string) (string, error) {
	var res string
	for _, s := range strs {
		res += s
	}
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

func send() error {
	server, err := startServer()
	if err != nil {
		return err
	}
	var concatTask tasks.Signature
	var initTasks = func() {
		concatTask = tasks.Signature{
			Name: "concat",
			Args: []tasks.Arg{{Type: "[]string", Value: []string{"foo", "bar"}}},
		}
	}
	initTasks()
	asyncResult, err := server.SendTask(&concatTask)
	if err != nil {
		return fmt.Errorf("Could not send task: %s", err.Error())
	}
	results, err := asyncResult.Get(time.Millisecond * 5)
	if err != nil {
		return fmt.Errorf("Getting task result failed with error: %s", err.Error())
	}
	log.INFO.Printf("concat([\"foo\", \"bar\"]) = %v\n", tasks.HumanReadableResults(results))
	return nil
}
