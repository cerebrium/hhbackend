package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// structures
type Color struct {
	_id string `json:"id,omitempty"`

	Color string `json:"color,omitempty"`

	Name string `json:"name,omitempty"`
}

// function for accessing the database
func GetMongoDbConnection() (*mongo.Client, error) {

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb+srv://admin:password1A@cyrptocluster.lmp7s.mongodb.net/sample_airbnb?retryWrites=true&w=majority"))

	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.Background(), readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}

	return client, nil
}

// function for connection to specific db and collection
func getMongoDbCollection(DbName string, CollectionName string) (*mongo.Collection, error) {

	client, err := GetMongoDbConnection()

	// handle errors
	if err != nil {
		return nil, err
	}

	collection := client.Database(DbName).Collection(CollectionName)

	return collection, nil
}

// function for finding a matching
func findOneColor(ColorId string) bool {
	// connect to the database
	collection, err := getMongoDbCollection("hhtest", "colors")
	if err != nil {
		// error in connection return error
		return true
	}

	// filter to only get one key
	var filter bson.M = bson.M{"color": ColorId}

	// make the results be in the correct format
	var results []bson.M

	// actually make the request using the cursor
	cur, err := collection.Find(context.Background(), filter)
	defer cur.Close(context.Background())

	// handle errors
	if err != nil {
		return true
	}

	// grab all of the results from the quesry
	cur.All(context.Background(), &results)

	// return found color or not
	if results == nil {
		return false
	} else {
		return true
	}
}

func main() {
	app := fiber.New()

	// allow for cors
	app.Use(cors.New())

	// basic get route
	app.Get("/", func(c *fiber.Ctx) error {
		// connect to the database
		collection, err := getMongoDbCollection("hhtest", "colors")
		if err != nil {
			// error in connection return error
			return c.Status(500).Send([]byte(err.Error()))
		}

		// filter
		var filter bson.M = bson.M{}

		// make the results be in the correct format
		var results []bson.M

		// actually make the request using the cursor
		cur, err := collection.Find(context.Background(), filter, options.Find())
		defer cur.Close(context.Background())

		// handle errors
		if err != nil {
			return c.Status(500).Send([]byte(err.Error()))
		}

		// grab all of the results from the quesry
		cur.All(context.Background(), &results)

		// handle errors
		if results == nil {
			return c.Status(404).SendString("not Found")
		}

		// send in json format the results
		json, _ := json.Marshal(results)
		return c.Send(json)
	})

	// get color by id
	app.Get("/colorid/:id?", func(c *fiber.Ctx) error {
		// connect to the database
		collection, err := getMongoDbCollection("hhtest", "colors")
		if err != nil {
			// error in connection return error
			return c.Status(500).Send([]byte(err.Error()))
		}

		// filter to only get one key
		var filter bson.M = bson.M{}
		if c.Params("id") != "" {
			_id := c.Params("id")
			filter = bson.M{"_id": _id}
		}

		// make the results be in the correct format
		var results []bson.M

		// actually make the request using the cursor
		cur, err := collection.Find(context.Background(), filter)
		defer cur.Close(context.Background())

		// handle errors
		if err != nil {
			return c.Status(500).Send([]byte(err.Error()))
		}

		// grab all of the results from the quesry
		cur.All(context.Background(), &results)

		// handle errors
		if results == nil {
			return c.Status(404).SendString("not Found")
		}

		// send in json format the results
		json, _ := json.Marshal(results)
		return c.Send(json)
	})

	// get color by color
	app.Get("/color/:color?", func(c *fiber.Ctx) error {
		// connect to the database
		collection, err := getMongoDbCollection("hhtest", "colors")
		if err != nil {
			// error in connection return error
			return c.Status(500).Send([]byte(err.Error()))
		}

		// filter to only get one key
		var filter bson.M = bson.M{}
		if c.Params("color") != "" {
			color := c.Params("color")
			filter = bson.M{"color": color}
		}

		// make the results be in the correct format
		var results []bson.M

		// actually make the request using the cursor
		cur, err := collection.Find(context.Background(), filter)
		defer cur.Close(context.Background())

		// handle errors
		if err != nil {
			return c.Status(500).Send([]byte(err.Error()))
		}

		// grab all of the results from the quesry
		cur.All(context.Background(), &results)

		// handle errors
		if results == nil {
			return c.Status(404).SendString("not Found")
		}

		// send in json format the results
		json, _ := json.Marshal(results)
		return c.Send(json)
	})

	// add a color
	app.Post("/addcolor", func(c *fiber.Ctx) error {

		//  declare struct
		var color Color

		// get the request body
		json.Unmarshal([]byte(c.Body()), &color)

		if findOneColor(color.Color) {
			return c.Status(500).SendString("color already in database")
		} else {
			// connect to the database
			collection, err := getMongoDbCollection("hhtest", "colors")
			if err != nil {
				// error in connection return error
				return c.Status(500).Send([]byte(err.Error()))
			}

			// check to make sure they submitted a color
			if color.Color != "" && color.Name != "" {
				// insert color
				res, err := collection.InsertOne(context.Background(), color)
				if err != nil {
					return c.Status(500).Send([]byte(err.Error()))
				}

				response, _ := json.Marshal(res)
				return c.Send(response)
			} else {
				fmt.Println(color)
				return c.Status(500).SendString("incorrect post content")
			}
		}
	})

	port := ":" + os.Getenv("PORT")

	if port == "" {
		port = "5000"
	}
	app.Listen(port)
}
