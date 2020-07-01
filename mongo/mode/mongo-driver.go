/*
Refer 
	https://github.com/tfogo/mongodb-go-tutorial/blob/master/main.go
Run
	Connected to MongoDB(mongodb://localhost:27017/)
	Insert: Person({Name:John Phone:13510987645},{Name:Tom Phone:13973521568})
	Found:  &{Tom 13973521568}
	-------------------------
	Insert: Person({Name:John Phone:13510987645},{Name:Tom Phone:13973521568})	
	Found:  &{Tom 13973521568}
	-------------------------
	......
*/
package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Person struct {
	Name, Phone string
}

var globalClientInstance *mongo.Client
var mu sync.RWMutex

func Run(client *mongo.Client) {
	if client == nil {
		return
	}

	// write
	collection := client.Database("test").Collection("people")
	John := Person{"John", "13510987645"}
	Tom := Person{"Tom", "13973521568"}
	peoples := []interface{}{John, Tom}
	_, err := collection.InsertMany(context.TODO(), peoples)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("Insert: Person(%+v,%+v)\n", John, Tom)

	// read
	filter := bson.D{{"name", "Tom"}}
	findOptions := options.Find()
	findOptions.SetLimit(2)
	var results []*Person
	cur, err := collection.Find(context.TODO(), filter, findOptions)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	for cur.Next(context.TODO()) { // Iterate through the cursor
		var elem Person
		err := cur.Decode(&elem)
		if err == nil {
			results = append(results, &elem)
		}
	}
	for _, elem := range results {
		fmt.Println("Found: ", elem)
	}
	cur.Close(context.TODO()) // Close the cursor once finished

	// delete
	_, err = collection.DeleteMany(context.TODO(), bson.D{{}})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("-------------------------")
}

func getClientInstance(url string) (client *mongo.Client) {
	if globalClientInstance == nil {
		mu.Lock()
		defer mu.Unlock()
		if globalClientInstance == nil {
			clientOptions := options.Client().ApplyURI(url)                        // Set client options
			globalClientInstance, _ = mongo.Connect(context.TODO(), clientOptions) // Connect to MongoDB
			err := globalClientInstance.Ping(context.TODO(), nil)                  // Check the connection
			if err == nil {
				fmt.Printf("Connected to MongoDB(%s)\n", url)
			}
		}
	}
	return globalClientInstance
}

func main() {
	url := "mongodb://localhost:27017/?maxIdleTimeMS=10000&maxPoolSize=3"
	client := getClientInstance(url)
	if client == nil {
		fmt.Println("getClientInstance failed!")
		return
	}
	for {
		go Run(client)
		time.Sleep(time.Second * 3)

	}
	err := globalClientInstance.Disconnect(context.TODO())
	if err != nil {
		fmt.Println(err)
	}
}
