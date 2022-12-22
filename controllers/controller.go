package controllers

import (
	"context"
	"fmt"
	"go-ecommerce/database"
	"go-ecommerce/models"
	generate "go-ecommerce/token"

	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var (
	UserCollection    *mongo.Collection = database.UserData(database.Client, "Users")
	ProductCollection *mongo.Collection = database.ProductData(database.Client, "Products")
	Validate                            = validator.New()
)

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

func VerifyPassword(userPassword, givenPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(givenPassword), []byte(userPassword))
	valid := true
	msg := ""

	if err != nil {
		msg = "Login or Password is incorrect"
		valid = false
	}
	return valid, msg
}

func Signup() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		//check if email address exists
		count, err := UserCollection.CountDocuments(ctx, bson.M{"email": user.Email})

		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		}

		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User already exists"})
		}

		//check if phone exists
		count, err = UserCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "This phone number is already in use"})
		}

		password := HashPassword(*user.Password)
		user.Password = &password

		user.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.UserId = user.ID.Hex()

		token, refreshToken, _ := generate.TokenGenerator(*user.Email, *user.FirstName, *user.LastName, user.UserId)
		user.Token = &token
		user.RefreshToken = &refreshToken
		user.UserCart = make([]models.UserProduct, 0)
		user.AddressDetails = make([]models.Address, 0)
		user.OrderStatus = make([]models.Order, 0)

		if _, insertErr := UserCollection.InsertOne(ctx, user); insertErr != nil {
			fmt.Println("Error in creating user: ", insertErr)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "The user did not get created"})

		}
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user, foundUser models.User
		// put the user details captured in the request context [c] in user model
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := UserCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		//verify password
		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		defer cancel()

		if !passwordIsValid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			fmt.Println("Error verifying password: ", msg)
			return

		}
		token, refreshToken, _ := generate.TokenGenerator(*foundUser.Email, *foundUser.FirstName, *foundUser.LastName, foundUser.UserId)
		defer cancel() //WHY DEFER CANCEL ALL THE TIME??

		// update user tokens
		generate.UpdateAllTokens(token, refreshToken, foundUser.UserId)
		c.JSON(http.StatusFound, foundUser)
	}
}

func ProductViewerAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var products models.Product
		defer cancel()

		if err := c.BindJSON(&products); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		products.ProductId = primitive.NewObjectID()
		if _, err := ProductCollection.InsertOne(ctx, products); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "not inserted"})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, "successfully added")
	}
}

func SearchProduct() gin.HandlerFunc {
	return func(c *gin.Context) {
		var productList []models.Product
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// pass empty braces to return all collections
		cursor, err := ProductCollection.Find(ctx, bson.D{{}})

		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "Soemtiong went wrong, please try again later")
			return
		}

		err = cursor.All(ctx, &productList)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		defer cursor.Close(ctx)

		if err = cursor.Err(); err != nil {
			log.Println(err)
			c.IndentedJSON(http.StatusBadRequest, "Invalid")
			return
		}
		defer cancel()
		c.IndentedJSON(http.StatusOK, productList)
	}
}

func SearchProductByQuery() gin.HandlerFunc {
	return func(c *gin.Context) {
		var searchProducts []models.Product
		queryParam := c.Query("name")

		// check if the queryParam is emopty
		if queryParam == "" {
			log.Println("query is empty")

			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid search index"})
			c.Abort()
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		searchQueryDb, err := ProductCollection.Find(ctx, bson.M{"productName": bson.M{"$regex": queryParam}})

		if err != nil {
			c.IndentedJSON(404, "Something went wrong while fetching product")
			return
		}

		err = searchQueryDb.All(ctx, &searchProducts)
		if err != nil {
			log.Println(err)
			c.IndentedJSON(400, "Invalid")
			return
		}
		defer searchQueryDb.Close(ctx)

		if err := searchQueryDb.Err(); err != nil {
			log.Println(err)
			c.IndentedJSON(400, "Invalid request")
			return
		}

		defer cancel()
		c.IndentedJSON(200, searchProducts)
	}
}
