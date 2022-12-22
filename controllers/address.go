package controllers

import (
	"context"
	"fmt"
	"go-ecommerce/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func AddAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Query("id")

		if id == "" {
			log.Println("id is empty")

			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid id"})
			c.Abort()
			return
		}
		addressId, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.IndentedJSON(500, "Internal Server Error")
			return
		}

		//find out how many addresses the user has
		var addresses models.Address

		addresses.AddressId = primitive.NewObjectID()

		if err = c.BindJSON(&addressId); err != nil {
			c.IndentedJSON(http.StatusNotAcceptable, err.Error())
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// create aggregation query
		// match finds the particular user with user id and its data
		matchFilter := bson.D{{Key: "$match", Value: bson.D{primitive.E{Key: "_id", Value: addressId}}}}
		// unwind stage will unwind the userCart data from a closed array to something that can be processed in go
		unwind := bson.D{{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$address"}}}}

		// grouping stage groups all the values with the help of id and find the total price
		grouping := bson.D{{Key: "$group", Value: bson.D{primitive.E{Key: "$_id", Value: "$addressId"}, {Key: "count", Value: bson.D{primitive.E{Key: "$sum", Value: 1}}}}}}
		pointCursor, err := UserCollection.Aggregate(ctx, mongo.Pipeline{matchFilter, unwind, grouping})

		if err != nil {
			c.IndentedJSON(500, "Internal Server Error")
		}

		var addressInfo []bson.M
		if err = pointCursor.All(ctx, &addressInfo); err != nil {
			panic(err)
		}

		var size int32
		for _, addressNo := range addressInfo {
			count := addressNo["count"]
			size = count.(int32)
		}

		if size < 2 {
			filter := bson.D{primitive.E{Key: "_id", Value: addressId}}
			update := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "address", Value: addresses}}}}
			if _, err := UserCollection.UpdateOne(ctx, filter, update); err != nil {
				fmt.Println(err)
			}

		} else {
			c.IndentedJSON(400, "Not allowed as have 2 addresses already")
		}
		defer cancel()
		ctx.Done()

	}
}

func EditHomeAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Query("id")

		if id == "" {
			log.Println("id is empty")

			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid search index"})
			c.Abort()
			return
		}

		userId, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.IndentedJSON(500, "Internal Server Error")
			return
		}

		var editAddress models.Address
		if err := c.BindJSON(&editAddress); err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		filter := bson.D{primitive.E{Key: "_id", Value: userId}}
		// update address values - home address is at 0th index
		update := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key: "address.0.house", Value: editAddress.House}, {Key: "address.0.street", Value: editAddress.Street}, {Key: "address.0.city", Value: editAddress.City}, {Key: "address.0.pinCode", Value: editAddress.PinCode}}}}

		_, err = UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.IndentedJSON(500, "Something went wrong")
			return
		}
		defer cancel()
		ctx.Done()
		c.IndentedJSON(200, "Successfully update dthe home address")
	}
}

func EditWorkAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Query("id")

		if id == "" {
			log.Println("id is empty")

			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid search index"})
			c.Abort()
			return
		}

		userId, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.IndentedJSON(500, "Internal Server Error")
			return
		}

		var editAddress models.Address
		if err := c.BindJSON(&editAddress); err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		filter := bson.D{primitive.E{Key: "_id", Value: userId}}
		// update address values - workaddress is at 0th index
		update := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key: "address.1.house", Value: editAddress.House}, {Key: "address.1.street", Value: editAddress.Street}, {Key: "address.1.city", Value: editAddress.City}, {Key: "address.1.pinCode", Value: editAddress.PinCode}}}}

		_, err = UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.IndentedJSON(500, "Something went wrong")
			return
		}
		defer cancel()
		ctx.Done()
		c.IndentedJSON(200, "Successfully updated the work address")
	}
}

func DeleteAddress() gin.HandlerFunc {

	return func(c *gin.Context) {
		id := c.Query("id")

		if id == "" {
			log.Println("id is empty")

			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid search index"})
			c.Abort()
			return
		}

		addresses := make([]models.Address, 0)
		userId, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.IndentedJSON(500, "Internal Server Error")
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		filter := bson.D{primitive.E{Key: "_id", Value: userId}}
		update := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key: "address", Value: addresses}}}} // set address to nil struct

		_, err = UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.IndentedJSON(404, "Wrong command")
			return
		}
		defer cancel()
		ctx.Done()
		c.IndentedJSON(200, "Successfully deleted")

	}
}
