package actions

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"time"

	"github.com/gobuffalo/buffalo"
)

// HomeHandler is a default handler to serve up
// a home page.
func RetrieveData(c buffalo.Context) error {

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://username:password@10.97.103.216:27017"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	//Connection to MDB
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	collection := client.Database("project-atlas").Collection("volume")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	var volumes []bson.M
	if err = cursor.All(ctx, &volumes); err != nil {
		log.Fatal(err)
	}
	fmt.Println(volumes)
	for _, volume := range volumes {
		fmt.Println(volume["podname"])
		fmt.Println(volume["volumename"])
	}

	return c.Render(http.StatusOK, r.HTML("retrieve_data.plush.html"))
}
