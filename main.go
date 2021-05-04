package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// structures
type Color struct {
	Id string `json:"_id,omitempty"`

	Color string `json:"color,omitempty"`

	Name string `json:"name,omitempty"`
}

// structure for the color object thats sortable
type ColorObject struct {
	Id string

	Color string

	Name string

	Blue int16

	Red int16

	Green int16

	Chroma float64

	Sat float64

	Val float64

	Luma float64

	Hue float64
}

// use godot package to load/read the .env file and
// return the value of the key
func goDotEnvVariable(key string) string {

	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}

// array maker helper
func createColor(color Color) ColorObject {
	// make the object to return
	m := new(ColorObject)

	// take the color and generate red, blue and green
	r, err := strconv.ParseInt(color.Color[1:3], 16, 64)
	g, err := strconv.ParseInt(color.Color[3:5], 16, 64)
	b, err := strconv.ParseInt(color.Color[5:7], 16, 64)

	// check for error
	if err != nil {
		fmt.Println(err)
	}

	// divide and convert type
	r2 := float64(r) / float64(255)
	g2 := float64(g) / float64(255)
	b2 := float64(b) / float64(255)

	// instantiate values for min and max
	min := float64(1)
	max := float64(0)

	// get max and min for the chroma
	colorArray := []float64{r2, g2, b2}

	// loop through array and determine min an max
	for _, color := range colorArray {
		if color > max {
			max = color
		}

		if color < min {
			min = color
		}
	}

	// HSV values of hex color
	chr := max - min
	hue := float64(0)
	val := max
	sat := float64(0)

	if val > 0 {
		// Calculate Saturation only if Value isn't 0
		sat = chr / val
		if sat > 0 {
			if r2 == max {
				hue = 60 * (((g2 - min) - (b2 - min)) / chr)
				if hue < 0 {
					hue += 360
				}
			} else if g2 == max {
				hue = 120 + 60*(((b2-min)-(r2-min))/chr)
			} else if b2 == max {
				hue = 240 + 60*(((r2-min)-(g2-min))/chr)
			}
		}
	}

	// set the values to the map to return
	m.Id = color.Id
	m.Name = color.Name
	m.Color = color.Color
	m.Blue = int16(b)
	m.Red = int16(r)
	m.Green = int16(g)
	m.Chroma = chr
	m.Hue = hue
	m.Luma = 0.3*r2 + 0.59*g2 + 0.11*b2
	m.Sat = sat
	m.Val = val

	return *m
}

// function to sort colors
func createSortableArray(array []Color) []ColorObject {
	// instantiate the array to return
	var finalArray []ColorObject

	// loop through the array and append each new color
	for _, color := range array {
		finalArray = append(finalArray, createColor(color))
	}

	// return the array
	return finalArray
}

// function for accessing the database
func GetMongoDbConnection() (*mongo.Client, error) {

	// grab the url for the database
	dotenv := goDotEnvVariable("MONGO_STRING")

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(dotenv))

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
	// load the env file
	godotenv.Load()

	// instantiate instance of fiber
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

		// turn the data into json
		json, err := json.Marshal(results)

		// handle errors
		if err != nil {
			return c.Status(500).Send([]byte(err.Error()))
		}

		// send the data
		return c.Send(json)
	})

	app.Get("/colors", func(c *fiber.Ctx) error {
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
		jsonData, err := json.Marshal(results)

		// handle errors
		if err != nil {
			return c.Status(500).Send([]byte(err.Error()))
		}

		// instantiate type array of Color
		var data []Color

		// unmarshal the []byte into usable format
		json.Unmarshal(jsonData, &data)

		// remarshal back to []byte
		jsonFinal, err := json.Marshal(createSortableArray(data))

		// handle errors
		if err != nil {
			return c.Status(500).Send([]byte(err.Error()))
		}

		// send the converted data
		return c.Send(jsonFinal)
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

	// allow for heroku to set port
	port := ":" + os.Getenv("PORT")

	if port == "" {
		port = "5000"
	}
	app.Listen(port)
}
