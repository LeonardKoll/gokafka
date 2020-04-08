package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/segmentio/kafka-go"
)

type uwa struct {
	Name   string
	Phone  string
	Street string
	Zip    string
	City   string
}

type namekey struct {
	Name string
}

type msg struct {
	Old uwa
	New uwa
}

type config struct {
	AwsTable  string
	Partition int
	KafkaHost string
	Offset    int64
	Topic     string
}

func main() {

	///////////////////////////////////////
	// Get Config Ready

	absPath, _ := filepath.Abs("./config.json")
	dat, _ := ioutil.ReadFile(absPath)
	var config config
	json.Unmarshal(dat, &config)

	///////////////////////////////////////
	// Get Kafka Ready

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{config.KafkaHost},
		Topic:     config.Topic,
		Partition: config.Partition,
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MB
	})
	r.SetOffset(config.Offset)
	defer r.Close()

	///////////////////////////////////////
	// Get DynamoDB ready

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := dynamodb.New(sess)

	///////////////////////////////////////
	// Handle Messages

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			break
		}

		write(svc, config.AwsTable, m.Value)
	}

}

func write(svc *dynamodb.DynamoDB, awsTable string, jsonstring []byte) {

	var parsed msg
	json.Unmarshal(jsonstring, &parsed)

	if !empty(parsed.New) { // Create or Replace

		av, err := dynamodbattribute.MarshalMap(parsed.New)
		if err != nil {
			fmt.Println("Got error marshalling map:")
			fmt.Println(err.Error())
		}

		input := &dynamodb.PutItemInput{
			Item:      av,
			TableName: aws.String(awsTable),
		}

		_, err = svc.PutItem(input)
		if err != nil {
			fmt.Println("Got error calling PutItem:")
			fmt.Println(err.Error())
		}

	} else { // Delete

		var key = namekey{Name: parsed.Old.Name}
		av, err := dynamodbattribute.MarshalMap(key)
		if err != nil {
			panic(fmt.Sprintf("failed to DynamoDB marshal Record, %v", err))
		}

		input := &dynamodb.DeleteItemInput{
			Key:       av,
			TableName: aws.String(awsTable),
		}

		_, err = svc.DeleteItem(input)
		if err != nil {
			fmt.Println("Got error deleting Item:")
			fmt.Println(err.Error())
		}

	}

}

func empty(tocheck uwa) bool {
	switch {
	case tocheck.City != "":
		return false
	case tocheck.Name != "":
		return false
	case tocheck.Phone != "":
		return false
	case tocheck.Street != "":
		return false
	case tocheck.Zip != "":
		return false
	}
	return true
}
