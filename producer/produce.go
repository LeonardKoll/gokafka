package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/segmentio/kafka-go"
)

type config struct {
	Topic     string
	Partition int
	KafkaHost string
	WatchFile string
}

func main() {

	///////////////////////////////////////
	// Get Config Ready

	absPath, _ := filepath.Abs("./config.json")
	dat, _ := ioutil.ReadFile(absPath)
	var config config
	json.Unmarshal(dat, &config)

	///////////////////////////////////////
	// Get File-Watcher Ready

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("ERROR", err)
	}
	defer watcher.Close()

	if err := watcher.Add(config.WatchFile); err != nil {
		fmt.Println("ERROR", err)
	}

	///////////////////////////////////////
	// Get Kafka Ready

	conn, err := kafka.DialLeader(context.Background(), "tcp", config.KafkaHost, config.Topic, config.Partition)
	if err != nil {
		panic("failed to connect to kafka")
	}
	defer conn.Close()
	conn.SetWriteDeadline(time.Now().Add(3600 * time.Second))

	///////////////////////////////////////
	// Process file changes
	for {
		select {
		case event, _ := <-watcher.Events:

			// Send
			dat, _ := ioutil.ReadFile(config.WatchFile)
			var woLBR = removeLBR(string(dat))
			fmt.Println(woLBR)
			_, err = conn.WriteMessages(
				kafka.Message{Value: []byte(woLBR)},
			)
			if err != nil {
				fmt.Print(err)
			}
			fmt.Println("sent: ", event)

		case err, _ := <-watcher.Errors:
			fmt.Println("error:", err)
		}
	}
}

func removeLBR(text string) string {
	re := regexp.MustCompile(`\x{000D}\x{000A}|[\x{000A}\x{000B}\x{000C}\x{000D}\x{0085}\x{2028}\x{2029}]`)
	return re.ReplaceAllString(text, ``)
}
