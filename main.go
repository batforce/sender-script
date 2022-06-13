package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

type WorkRequest struct {
	RequestType   RequestType `json:"requestType"`
	Kit           Kit         `json:"kit"`
	RepositoryUrl string      `json:"repositoryUrl"`
	Ref           string      `json:"ref"`
	Variables     []Variable  `json:"variables"`
	Hash          string      `json:"hash"`
}

type Kit struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type RequestType string

const (
	PreCompile RequestType = "precompile"
	Compile    RequestType = "compile"
	Deploy     RequestType = "deploy"
)

type Buildkit struct {
	Url       string
	Dir       string
	Variables []Variable
}

type VariableType string

const (
	StringVariable VariableType = "string"
	SecretVariable VariableType = "secret"
)

type Variable struct {
	Key   string       `json:"key"`
	Value string       `json:"value"`
	Type  VariableType `json:"type"`
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"bat_worker_queue", // name
		true,               // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	failOnError(err, "Failed to declare a queue")

	body := bodyFrom(os.Args)

	request := &WorkRequest{
		RepositoryUrl: "https://github.com/batforce/simple",
		Ref:           "hgehehehe",
		RequestType:   Compile,
		Kit: Kit{
			Name: "buildkit-docker-main",
			Url:  "https://github.com/batforce/buildkit-docker",
		},
		Hash: "2356c3923eb886f5d767d32e76d1b5e907dc02f8",
		Variables: []Variable{
			{
				Key:   "APP_NAME",
				Value: "simple-app-1456",
				Type:  StringVariable,
			},
			{
				Key:   "DOCKER_REGISTRY",
				Value: "docker-registry:5006",
				Type:  StringVariable,
			},
		},
	}
	content, _ := json.Marshal(request)

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         content,
		})
	failOnError(err, "Failed to publish a message")

	log.Printf(" [x] Sent %s", body)
}

func bodyFrom(args []string) string {
	var s string
	if (len(args) < 2) || os.Args[1] == "" {
		s = "hello"
	} else {
		s = strings.Join(args[1:], " ")
	}
	return s
}
