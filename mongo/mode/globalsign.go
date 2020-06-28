package main

import (
	"fmt"
	"time"
  
	"github.com/globalsign/mgo"      //"gopkg.in/mgo.v2"
	"github.com/globalsign/mgo/bson" //"gopkg.in/mgo.v2/bson"
)

type Person struct {
	Name  string
	Phone string
}

func getNewSession() (session *mgo.Session, err error) {
	mongoURL := "mongodb://localhost:27017/?maxPoolSize=3&maxIdleTimeMS=50000"
	dailInfo, err := mgo.ParseURL(mongoURL)
	if err != nil {
		fmt.Printf("failed to parse to mongoDB url,dailInfo(%v) error %s\n",dailInfo, err.Error())
		return nil, err
	}
	mongoSession, err := mgo.Dial(mongoURL)
	if err != nil {
		fmt.Printf("failed to connect to mongoDB, error %s\n", err.Error())
		return nil, err
	}
	mongoSession.SetMode(mgo.Strong, true) //mgo.Eventual,mgo.Monotonic,mgo.Strong,mgo.Primary,mgo.PrimaryPreferred...
	fmt.Printf("successfully connect to mongoDB(%s)\n", mongoURL)
	return mongoSession.Copy(), nil
}

func Run() {
	session, err := getNewSession()
	if err != nil {
		return
	}
	defer session.Close()
  
	//write
	c := session.DB("test").C("people")
	err = c.Insert(&Person{"John", "13618302576"}, &Person{"Tom", "135109867468"})
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("Insert : Person OK!\n")
  
	//read
	result := Person{}
	err = c.Find(bson.M{"name": "Tom"}).One(&result)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("Find: Person{Name:%s,Phone:%s}\n", result.Name, result.Phone)
}

func main() {
	for {
		go Run()
		time.Sleep(time.Second * 3)
	}
}
