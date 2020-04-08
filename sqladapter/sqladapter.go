package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/segmentio/kafka-go"
)

type address struct {
	gorm.Model
	Street string
	Zip    string
	City   string
	UserID uint
}

type user struct {
	gorm.Model
	Name    string
	Phone   string
	Address address
}

type uwa struct {
	Name   string
	Phone  string
	Street string
	Zip    string
	City   string
}

type msg struct {
	Old uwa
	New uwa
}

type config struct {
	DbConnection string
	Topic        string
	Partition    int
	KafkaHost    string
	Offset       int64
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
	// Get Postgres Ready

	db, err := gorm.Open("postgres", config.DbConnection)
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()
	db.AutoMigrate(&address{})
	db.AutoMigrate(&user{})

	///////////////////////////////////////
	// Handle Messages

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			break
		}
		write(db, m.Value)
	}

}

func write(db *gorm.DB, jsonstring []byte) {
	var parsed msg
	json.Unmarshal(jsonstring, &parsed)

	switch {
	case empty(parsed.New): // Delete
		var todelete user
		db.First(&todelete, "name = ?", parsed.Old.Name)
		db.Delete(&todelete)
	case empty(parsed.Old): // Create
		var tocreate user
		convert(parsed.New, &tocreate)
		db.Create(&tocreate)
	default: // Update
		var toupdate user
		db.Set("gorm:auto_preload", true).First(&toupdate, "name = ?", parsed.Old.Name) // Autopreload makes sure that Address is initializied, otherwise it would be blank.
		convert(parsed.New, &toupdate)
		db.Model(&toupdate).Updates(toupdate)
	}
}

func convert(toconvert uwa, destination *user) {

	(*destination).Name = toconvert.Name
	(*destination).Phone = toconvert.Phone
	(*destination).Address.Street = toconvert.Street
	(*destination).Address.Zip = toconvert.Zip
	(*destination).Address.City = toconvert.City

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
