package database

import (
	"context"
	"errors"
	"go-ecommerce/models"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrCantFindProduct    = errors.New("cant find product")
	ErrCantDecodeProduct  = errors.New("cant decode product")
	ErrUserIdIsNotValid   = errors.New("user id is not valid")
	ErrCantUpdateUser     = errors.New("cant update user")
	ErrCantRemoveItemCart = errors.New("cant remove item from cart")
	ErrCantGetItem        = errors.New("cant get item")
	ErrCantBuyCartItem    = errors.New("cant buy cart item")
)

func AddToCart(ctx context.Context, prodCollection, userCollection *mongo.Collection, productId string) error {
	searchFromDb, err := prodCollection.Find(ctx, bson.M{"_id": productId})
	if err != nil {
		log.Println(err)
		return ErrCantFindProduct
	}

	var productCart []models.UserProduct
	err = searchFromDb.All(ctx, &productCart)
	if err != nil {
		log.Println(err)
		return ErrCantDecodeProduct
	}

	id, err := primitive.ObjectIDFromHex(productId)
	if err != nil {
		log.Println(err)
		return ErrUserIdIsNotValid
	}

	filter := bson.D{primitive.E{Key: "_id", Value: id}}
	update := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "userCart", Value: bson.D{{Key: "$each", Value: productCart}}}}}}

	_, err = userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Println(err)
		return ErrCantUpdateUser
	}
	return nil

}

func RemoveItemFromCart(ctx context.Context, prodCollection, userCollection *mongo.Collection, productId primitive.ObjectID, userId string) error {
	id, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		log.Println(err)
		return ErrUserIdIsNotValid
	}
	filter := bson.D{primitive.E{Key: "_id", Value: id}}
	// remove a particular item from the cart list of the user
	update := bson.M{"$pull": bson.M{"userCart": bson.M{"_id": productId}}}
	_, err = userCollection.UpdateMany(ctx, filter, update)

	if err != nil {
		log.Println(err)
		return ErrCantRemoveItemCart
	}

	return nil
}

// func GetItemFromCart() gin.HandlerFunc {

// }
func BuyItemFromCart(ctx context.Context, userCollection *mongo.Collection, userId string) error {
	// fetch the cart of the user
	// find the cart total
	// create an order with the items
	// add order to the user collection
	// add items in the cart to list of user orders
	// empty the cart

	id, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		log.Println(err)
		return ErrUserIdIsNotValid
	}

	var getCartItems models.User
	var orderCart models.Order

	orderCart.OrderId = primitive.NewObjectID()
	orderCart.OrderedAt = time.Now()
	orderCart.OrderCart = make([]models.UserProduct, 0)
	orderCart.PaymentMethod.COD = true

	// add all user products in cart
	// unwind gives us access to each value in cart
	// path tells us what we want to unwind
	unwind := bson.D{{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$userCart"}}}}
	//group by id and find the total which is the sum of all product prices in cart
	grouping := bson.D{{Key: "$group", Value: bson.D{primitive.E{Key: "_id", Value: "$_id"}, {Key: "total", Value: bson.D{primitive.E{Key: "$sum", Value: "$userCart.price"}}}}}}
	currResults, err := userCollection.Aggregate(ctx, mongo.Pipeline{unwind, grouping})
	ctx.Done()

	if err != nil {
		panic(err)
	}

	// create an order with the items
	var getUserCart []bson.M
	if err = currResults.All(ctx, &getUserCart); err != nil {
		panic(err)
	}

	var totalPrice int32

	for _, userItem := range getUserCart {
		price := userItem["total"]
		totalPrice += price.(int32)
	}

	orderCart.Price = int(totalPrice)

	filter := bson.D{primitive.E{Key: "_id", Value: id}}

	update := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "orders", Value: orderCart}}}}
	if _, err = userCollection.UpdateMany(ctx, filter, update); err != nil {
		log.Println(err)
	}

	if err = userCollection.FindOne(ctx, bson.D{primitive.E{Key: "_id", Value: id}}).Decode(&getCartItems); err != nil {
		log.Println(err)
	}

	// add the items in the cart to the order to make a list of orders he has ordered
	filter2 := bson.D{primitive.E{Key: "_id", Value: id}}
	update2 := bson.M{"$push": bson.M{"orders.$[].orderList": bson.M{"$each": getCartItems.UserCart}}}
	if _, err = userCollection.UpdateOne(ctx, filter2, update2); err != nil {
		log.Println(err)
	}

	// clear cart
	userCartEmpty := make([]models.UserProduct, 0)
	filter3 := bson.D{primitive.E{Key: "_id", Value: id}}
	update3 := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key: "userCart", Value: userCartEmpty}}}}

	if _, err = userCollection.UpdateOne(ctx, filter3, update3); err != nil {
		log.Println(err)
		return ErrCantBuyCartItem
	}
	return nil
}

func InstantBuy(ctx context.Context, prodCollection, userCollection *mongo.Collection, productId primitive.ObjectID, userId string) error {
	// create an order
	id, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		log.Println(err)
		return ErrUserIdIsNotValid
	}
	var productDetails models.UserProduct
	var ordersDetail models.Order

	ordersDetail.OrderId = primitive.NewObjectID()
	ordersDetail.OrderedAt = time.Now()
	ordersDetail.OrderCart = make([]models.UserProduct, 0)
	ordersDetail.PaymentMethod.COD = true

	// use the product id and find the product from the db
	if err = prodCollection.FindOne(ctx, bson.D{primitive.E{Key: "_id", Value: productId}}).Decode(&productDetails); err != nil {
		log.Println(err)
	}

	ordersDetail.Price = productDetails.Price

	filter := bson.D{primitive.E{Key: "_id", Value: id}}
	update := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "order", Value: ordersDetail}}}}
	if _, err = userCollection.UpdateOne(ctx, filter, update); err != nil {
		log.Println(err)
	}

	filter2 := bson.D{primitive.E{Key: "_id", Value: id}}
	update2 := bson.M{"$push": bson.M{"orders.$[].orderList": productDetails}}
	if _, err = userCollection.UpdateOne(ctx, filter2, update2); err != nil {
		log.Println(err)
	}

	return nil

}
