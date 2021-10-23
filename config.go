package main

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/yaml.v2"
)

type BotConf struct {
	Token  string `yaml:"token"`
	Prefix string `yaml:"prefix"`
	Uri    string `yaml:"uri"`
	BotID  string
}

var (
	conf           BotConf
	clientUser     *mongo.Client
	ctxUser        context.Context
	collectionUser *mongo.Collection
)

//will connect to database on user's collectionn
func ConnectToDatabaseUsers() error {
	ctxUser, _ := context.WithTimeout(context.TODO(), 10*time.Second)

	//try to connect
	clientOptions := options.Client().ApplyURI(conf.Uri)
	clientUser, err := mongo.Connect(ctxUser, clientOptions)
	if err != nil {
		return err
	}

	//check if connection is established
	err = clientUser.Ping(context.TODO(), nil)
	if err != nil {
		return err
	}

	//assign to the global variable "collection" the users' collection
	collectionUser = clientUser.Database("monkey-bot").Collection("typers")
	return nil
}

//read the config file and fill the config struct
func init() {
	dat, err := ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("error in init function: %v", err.Error())
	}
	err = yaml.Unmarshal([]byte(dat), &conf)
	if err != nil {
		log.Fatalf("error unmarshalling config: %v", err)
	}

	err = ConnectToDatabaseUsers()
	if err != nil {
		log.Fatalf("error opening database: %v", err)
	}
}
