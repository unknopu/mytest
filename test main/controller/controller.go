package controller

import (
	"context"
	"fmt"
	"log"
	"mongodb-go/cst"
	"mongodb-go/models"
	"net/http"
	"reflect"
	"time"

	// "reflect"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"

	// "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func connectDB() (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(cst.DatabaseEndPint)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		fmt.Println("\n+--------------------+ERROR while Mongo.connect():", err)
		log.Fatal(err)
	}
	return client, err
}

func Registoration(c *gin.Context) {
	var u models.User
	c.BindJSON(&u)
	if u.UserName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"err": "Username can not be empty",
		})
		return
	}
	u, e := AddUser(u)
	if e {
		c.JSON(http.StatusOK, gin.H{
			"NewUser": u,
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"err": "It's existing user",
		})
	}
}

func AddUser(user models.User) (models.User, bool) {
	client, _ := connectDB()

	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	col := client.Database("game").Collection("user")
	defer client.Disconnect(ctx)

	data := user
	// data.ID = primitive.NewObjectID()

	existing, _ := col.Find(ctx, bson.M{"username": user.UserName})
	var Filtered []bson.M
	if err := existing.All(ctx, &Filtered); err != nil {
		fmt.Println("Error filltering: ", err)
		return data, false
	}
	if Filtered != nil {
		fmt.Println("\n\nThis user is existing", Filtered)
		return data, false
	}

	data.MakeDefaultVal()

	result, insertErr := col.InsertOne(ctx, data)
	if insertErr != nil {
		fmt.Println("InsertOne Error: ", insertErr)
	} else {
		newID := result.InsertedID
		fmt.Println("InsertedOne newID: ", reflect.TypeOf(newID))
	}
	return data, true
}

func DisplayAllUser(c *gin.Context) {
	client, _ := connectDB()

	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	defer client.Disconnect(ctx)
	col := client.Database("game").Collection("user")

	cursor, err := col.Find(ctx, bson.D{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var users models.User
		if err = cursor.Decode(&users); err != nil {
			log.Fatal(err)
		}
		fmt.Println(users)
		c.JSON(http.StatusOK, gin.H{
			"username": users.UserName,
			"status":   cst.TextStatus[users.Status],
		})
	}
}

func Ranking(c *gin.Context) {
	client, _ := connectDB()

	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	defer client.Disconnect(ctx)
	col := client.Database("game").Collection("user")

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"w", -1}})
	cursor, err := col.Find(ctx, bson.D{}, findOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var users bson.M
		if err = cursor.Decode(&users); err != nil {
			log.Fatal(err)
		}
		// c.JSON(http.StatusOK, users["username"])
		c.HTML(http.StatusOK, "ranking.html", gin.H{
			"username": users["username"],
			"win":      users["w"],
		})
	}
}

func Search(c *gin.Context) {
	name := c.Query("name")
	client, _ := connectDB()
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	defer client.Disconnect(ctx)
	col := client.Database("game").Collection("user")

	filter := bson.M{"username": name}
	var user models.User

	err := col.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
	// c.Data(http.StatusOK, "text-html; charset=utf-8", []byte("done"))
	// c.HTML(http.StatusOK, "userInfo.html", gin.H{
	// 	"username": user.UserName,
	// 	"status": cst.Text[user.Status],
	// 	"win": user.W,
	// 	"lose": user.L,
	// 	"cname": user.LastMatch.UserName,
	// 	"cresult": user.LastMatch.Win,
	// })
}

func Challenge(c *gin.Context) {
	var record models.Record
	c.BindJSON(&record)

	client, _ := connectDB()
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	recordCollection := client.Database("game").Collection("record")
	userCollection := client.Database("game").Collection("user")
	defer client.Disconnect(ctx)

	// username validation
	var sender models.User
	var recver models.User
	err := userCollection.FindOne(ctx, bson.M{"username": record.SendBy}).Decode(&sender)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Sender does not exist"})
		return
	}
	err = userCollection.FindOne(ctx, bson.M{"username": record.SendTo}).Decode(&recver)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reciver does not exist"})
		return
	}

	// ever challenge or matche another?
	var prevRecord models.Record
	isDuplicate := recordCollection.FindOne(ctx, bson.M{"sendby": record.SendBy, "sendto": record.SendTo}).Decode(&prevRecord)
	isChallenged := recordCollection.FindOne(ctx, bson.M{"sendby": record.SendTo, "sendto": record.SendBy}).Decode(&prevRecord)
	// have sent but not match yet, notice waiting message
	if isDuplicate == nil && isChallenged != nil {
		c.JSON(http.StatusOK, gin.H{"message": "waiting for reply"})
		return
	}

	// new build challenge history
	var sender_history models.History
	sender_history.AddData(record.SendTo, 0)
	var recver_history models.History
	recver_history.AddData(record.SendBy, 0)

	// in case both challenging each other - start considering the result
	if isChallenged == nil {
		// in case - tie
		if prevRecord.Choise == record.Choise {
			result, _ := userCollection.UpdateOne(ctx,
				bson.M{"username": record.SendBy},
				bson.D{{"$set", bson.D{{"n", sender.N + 1}, {"status", 0}, {"lastmatch", sender_history}}}})
			fmt.Printf("Updated %v Documents!\n", result.ModifiedCount)
			result, _ = userCollection.UpdateOne(ctx,
				bson.M{"username": record.SendTo},
				bson.D{{"$set", bson.D{{"n", recver.N + 1}, {"status", 0}, {"lastmatch", recver_history}}}})
			fmt.Printf("Updated %v Documents!\n", result.ModifiedCount)
			delResult, _ := recordCollection.DeleteOne(ctx,
				bson.M{"sendby": record.SendTo, "sendto": record.SendBy})
			fmt.Printf("DeleteOne removed %v document(s)\n", delResult.DeletedCount)
			c.JSON(http.StatusOK, gin.H{"challenge status": "tie"})
			return
		}

		// other challenge result
		sender_history.AddData(record.SendTo, 1)
		recver_history.AddData(record.SendBy, 2)

		if !gameDecision(record.Choise, prevRecord.Choise) {
			sender, recver = recver, sender
			record, prevRecord = prevRecord, record
			sender_history.AddData(record.SendTo, 2)
			recver_history.AddData(record.SendBy, 1)
		}

		result, _ := userCollection.UpdateOne(ctx,
			bson.M{"username": record.SendBy},
			bson.D{{"$set", bson.D{
				{"w", sender.W + 1},
				{"status", 0},
				{"lastmatch", sender_history},
			}}})
		fmt.Printf("Updated %v Documents!\n", result.ModifiedCount)
		result, _ = userCollection.UpdateOne(ctx,
			bson.M{"username": record.SendTo},
			bson.D{{"$set", bson.D{
				{"l", recver.L + 1},
				{"status", 0},
				{"lastmatch", recver_history},
			}}})
		fmt.Printf("Updated %v Documents!\n", result.ModifiedCount)

		delResult, _ := recordCollection.DeleteOne(ctx,
			bson.M{"sendby": record.SendTo, "sendto": record.SendBy})
		fmt.Printf("DeleteOne removed %v document(s)\n", delResult.DeletedCount)
		c.JSON(http.StatusOK, gin.H{"challenge result": sender.UserName})
		return
	}

	// when challenge still not match
	// insert record to database, update user's status
	_, insertErr := recordCollection.InsertOne(ctx, record)
	if insertErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": insertErr})
	} else {
		c.JSON(http.StatusOK, gin.H{"newRecord": record})
		// update user data status
		result, _ := userCollection.UpdateOne(ctx,
			bson.M{"username": record.SendBy},
			bson.D{{"$set", bson.D{{"status", 2}}}})
		fmt.Printf("Updated %v Documents!\n", result.ModifiedCount)
		result, _ = userCollection.UpdateOne(ctx,
			bson.M{"username": record.SendTo},
			bson.D{{"$set", bson.D{{"status", 1}}}})
		fmt.Printf("Updated %v Documents!\n", result.ModifiedCount)
		return
	}
}

// sender lose=flase, sender win=true
func gameDecision(sender int, recv int) bool {
	sum := sender + recv
	if sum == 1 {
		return sender == 1
	} else if sum == 2 {
		return sender == 0
	} else if sum == 3 {
		return sender == 2
	}
	return false
}
