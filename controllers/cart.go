package controllers

import (
	"context"
	"errors"
	"go-ecommerce/database"
	"go-ecommerce/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Application struct {
	prodCollection *mongo.Collection
	userCollection *mongo.Collection
}

func NewApplication(prodCollection, userCollection *mongo.Collection) *Application {
	return &Application{prodCollection: prodCollection, userCollection: userCollection}
}

func (app *Application) AddToCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		// you need user who is adding and product to be added.
		productQueryId := c.Query("id")
		if productQueryId == "" {
			log.Println("Product ID is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("product ID is empty"))
			return
		}

		userQueryId := c.Query("userId")
		if userQueryId == "" {
			log.Println("User ID is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("user ID is empty"))
			return
		}

		// productId, err := primitive.ObjectIDFromHex(productQueryId)

		// if err != nil {
		// 	log.Println(err)
		// 	c.AbortWithStatus(http.StatusInternalServerError)
		// 	return
		// }

		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := database.AddToCart(ctx, app.prodCollection, app.userCollection, productQueryId)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
		}
		c.IndentedJSON(http.StatusOK, "Successfully added to cart")
	}
}

func (app *Application) RemoveItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		// you need user who is adding and product to be added.
		productQueryId := c.Query("id")
		if productQueryId == "" {
			log.Println("Product ID is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("product ID is empty"))
			return
		}

		userQueryId := c.Query("userId")
		if userQueryId == "" {
			log.Println("User ID is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("user ID is empty"))
			return
		}

		productId, err := primitive.ObjectIDFromHex(productQueryId)

		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = database.RemoveItemFromCart(ctx, app.prodCollection, app.userCollection, productId, userQueryId)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
		}
		c.IndentedJSON(http.StatusOK, "Successfully removed product from cart")
	}
}

func (app *Application) GetItemFromCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Query("id")

		if id == "" {
			log.Println("id is empty")

			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid id"})
			c.Abort()
			return
		}

		userId, _ := primitive.ObjectIDFromHex(id)

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var filledCart models.User

		if err := UserCollection.FindOne(ctx, bson.D{primitive.E{Key: "_id", Value: userId}}); err != nil {
			log.Println(err)
			c.IndentedJSON(500, "Not found")
			return
		}

		// create aggregation query
		// match finds the particular user with user id and its data
		filterMatch := bson.D{{Key: "$match", Value: bson.D{primitive.E{Key: "_id", Value: userId}}}}
		// unwind stage will unwind the userCart data from a closed array to something that can be processed in go
		unwind := bson.D{{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$userCart"}}}}

		// grouping stage groups all the values with the help of id and find the total price
		grouping := bson.D{{Key: "$group", Value: bson.D{primitive.E{Key: "$_id", Value: "$_id"}, {Key: "$sum", Value: "$userCart.price"}}}}
		pointCursor, err := UserCollection.Aggregate(ctx, mongo.Pipeline{filterMatch, unwind, grouping})

		if err != nil {
			log.Println(err)
		}

		var listing []bson.M
		// convert the values in pointCursor to listing
		if err = pointCursor.All(ctx, &listing); err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
		}

		for _, json := range listing {
			c.IndentedJSON(200, json["total"])
			c.IndentedJSON(200, filledCart.UserCart)
		}
		ctx.Done()
	}
}

func (app *Application) BuyFromCart() gin.HandlerFunc {
	return func(c *gin.Context) {

		userQueryId := c.Query("userId")
		if userQueryId == "" {
			log.Println("User ID is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("user ID is empty"))
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		err := database.BuyItemFromCart(ctx, app.userCollection, userQueryId)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
		}
		c.IndentedJSON(http.StatusOK, "Successfully placed the order")
	}
}

func (app *Application) InstantBuy() gin.HandlerFunc {
	return func(c *gin.Context) {
		// you need user who is adding and product to be added.
		productQueryId := c.Query("id")
		if productQueryId == "" {
			log.Println("Product ID is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("product ID is empty"))
			return
		}

		userQueryId := c.Query("userId")
		if userQueryId == "" {
			log.Println("User ID is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("user ID is empty"))
			return
		}

		productId, err := primitive.ObjectIDFromHex(productQueryId)

		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = database.InstantBuy(ctx, app.prodCollection, app.userCollection, productId, userQueryId)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
		}
		c.IndentedJSON(http.StatusOK, "Successfully placed the order")
	}
}
