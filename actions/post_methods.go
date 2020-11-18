package actions

import (
	"context"
	"github.com/gobuffalo/buffalo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"time"
)

type Person struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Firstname string             `json:"firstname,omitempty"`
	Lastname  string             `json:"lastname,omitempty"`
}

var client *mongo.Client

func PostFunc(c buffalo.Context) error {
	//															  mongodb://username:password@<your mongodb-service cluster-ip>:27017
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://username:password@10.103.9.137:27017"))
	if err != nil {
		log.Fatal(err)
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	//Connection to MDB
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	p := &Person{}
	if err := c.Bind(p); err != nil {
		return err
	}

	collection := client.Database("local").Collection("posts")
	_, err = collection.InsertOne(ctx, p)
	if err != nil {
		log.Fatal(err)
	}

	return c.Render(http.StatusOK, r.JSON(p))
}
