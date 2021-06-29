/*
Refer
	https://github.com/tfogo/mongodb-go-tutorial/blob/master/main.go
	https://github.com/hwholiday/learning_tools/blob/master/mongodb/mongo-go-driver/main.go
	https://github.com/mongodb/mongo-go-driver/tree/master/examples/documentation_examples
*/
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

type Trainer struct {
	Name string
	Age  int
	City string
}
type TrainerDB struct {
	Id   primitive.ObjectID `bson:"_id"`
	Name string             `bson:"name"`
	Age  int                `bson:"age"`
	City string             `bson:"city"`
}

func (t *Trainer) Connect(uri string) (client *mongo.Client) {
	want, err := readpref.New(readpref.PrimaryPreferredMode)
	if err != nil {
		log.Fatal(err)
	}
	wc := writeconcern.New(writeconcern.WMajority())
	readconcern.Majority()

	opt := options.Client().ApplyURI(uri)
	opt.SetLocalThreshold(3 * time.Second)     //只使用与mongo操作耗时秒数
	opt.SetMaxConnIdleTime(5 * time.Second)    //指定连接可以保持空闲的最大秒数,同url参数maxIdleTimeMS
	opt.SetMaxPoolSize(3)                      //使用最大的连接数,默认值是100
	opt.SetReadPreference(want)                //表示只从哪些节点读取数据
	opt.SetReadConcern(readconcern.Majority()) //指定查询应返回实例的最新数据确认为，已写入副本集中的大多数成员
	opt.SetWriteConcern(wc)                    //请求确认写操作传播到大多数mongod实例

	client, err = mongo.Connect(context.TODO(), opt) // Connect to MongoDB
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(context.TODO(), nil) // Check the connection
	if err != nil {
		log.Fatal(err)
	}
	return
}
func (t *Trainer) List(client *mongo.Client) (err error) {
	dbnames, err := client.ListDatabaseNames(context.TODO(), bson.M{})
	for _, dbname := range dbnames {
		db := client.Database(dbname)
		fmt.Println(dbname)
		collectionNames, _ := db.ListCollectionNames(context.TODO(), bson.M{})
		fmt.Println("\t", collectionNames)
	}
	return
}
func (t *Trainer) Count(client *mongo.Client) (err error) {
	dbNames, err := client.ListDatabaseNames(context.TODO(), bson.M{})
	for _, dbName := range dbNames {
		db := client.Database(dbName)
		fmt.Println(dbName)
		collectionNames, _ := db.ListCollectionNames(context.TODO(), bson.M{})
		for _, collectionName := range collectionNames {
			collection := client.Database(dbName).Collection(collectionName)
			count, _ := collection.CountDocuments(context.TODO(), bson.D{})
			fmt.Println("\t", collectionName, "(", count, ")")
		}
	}
	return
}
func (t *Trainer) Add(collection *mongo.Collection) (err error) {
	// Some dummy data to add to the Database
	ash := Trainer{"Ash", 10, "Pallet Town"}
	misty := Trainer{"Misty", 10, "Cerulean City"}
	brock := Trainer{"Brock", 15, "Pewter City"}

	// Insert a single document
	insertResult, err := collection.InsertOne(context.TODO(), ash)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted a single document: ", insertResult.InsertedID)

	// Insert multiple documents
	trainers := []interface{}{misty, brock}

	insertManyResult, err := collection.InsertMany(context.TODO(), trainers)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted multiple documents: ", insertManyResult.InsertedIDs)
	return
}
func (t *Trainer) Update(collection *mongo.Collection) (err error) {
	filter := bson.D{{"name", "Ash"}}
	update := bson.D{
		{"$inc", bson.D{{"age", 1}}},
	}
	updateResult, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Matched %v documents and updated %v documents.\n", updateResult.MatchedCount, updateResult.ModifiedCount)
	return
}
func (t *Trainer) Find(collection *mongo.Collection) (err error) {
	// Find a single document
	var result TrainerDB
	filter := bson.D{{"name", "Ash"}}
	err = collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Found a single document:", result.Id.Hex(), result.Name, result.Age, result.City)

	// Finding multiple documents returns a cursor
	findOptions := options.Find()
	findOptions.SetLimit(2) // 最大返回数
	var results []*Trainer
	cur, err := collection.Find(context.TODO(), bson.D{{}}, findOptions)
	if err != nil {
		log.Fatal(err)
	}
	for cur.Next(context.TODO()) { // Iterate through the cursor
		var elem Trainer
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, &elem)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	cur.Close(context.TODO()) // Close the cursor once finished
	fmt.Printf("Found multiple documents (array of pointers): %+v\n", results)

	return
}
func (t *Trainer) Delete(collection *mongo.Collection) (err error) {
	deleteResult, err := collection.DeleteMany(context.TODO(), bson.D{{}})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Deleted %v documents in the trainers collection\n", deleteResult.DeletedCount)
	return
}
func (t *Trainer) Disconnect(client *mongo.Client) (err error) {
	err = client.Disconnect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	return
}

func main() {
	var train Trainer
	var err error
	action := flag.String("a", "list", "The action of list|count|add|delete|update|find|drop")
	uri := flag.String("u", "mongodb://localhost:27017", "mongo uri ")
	db := flag.String("d", "test", "Database ")
	co := flag.String("c", "trainers", "Collection(Table)")
	flag.Parse()

	client := train.Connect(*uri)
	if client == nil {
		log.Fatal(err)
		return
	}
	defer train.Disconnect(client)

	switch *action {
	case "list":
		err = train.List(client)
	case "count":
		err = train.Count(client)
	case "add":
		collection := client.Database(*db).Collection(*co) //选择数据库和集合表
		err = train.Add(collection)
	case "update":
		collection := client.Database(*db).Collection(*co)
		err = train.Update(collection)
	case "find":
		collection := client.Database(*db).Collection(*co)
		err = train.Find(collection)
	case "delete":
		collection := client.Database(*db).Collection(*co)
		err = train.Delete(collection)
	case "drop":
		collection := client.Database(*db).Collection(*co)
		err = collection.Drop(context.TODO()) //删除数据库和集合表
	}
	if err != nil {
		log.Fatal(err)
		return
	}
}
