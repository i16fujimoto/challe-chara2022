package main

import(
	"context"
	"fmt"
	"time"

	"back-challe-chara2022/db"
	"back-challe-chara2022/entity/db_entity"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

)

// DBマイグレート
func main() {

	// DBの初期化
	db.InitDB()

	// test success

	// bearCollection init
	bearCollection := db.MongoClient.Database("insertDB").Collection("bears")
	bearId := primitive.NewObjectID()
	docBear := &db_entity.Bear{
		BearId: bearId,
		BearName: "kichiro",
        BearIcon: "img_dir/test.png",
        Detail: "テスト",
	}

	_, err1 := bearCollection.InsertOne(context.TODO(), docBear) // ここでMarshalBSON()される
	if err1 != nil {
		fmt.Println("Error inserting bear")
        panic(err1)
    } else {
		fmt.Println("Successfully inserting bear")
	}
	
	// bearToneCollection init
	bearToneCollection := db.MongoClient.Database("insertDB").Collection("bearTones")
	// 仮レスポンス
	response := [8]string{"そうだよね", "もう一回最初から教えてよ", "そこってどういうことなの？", 
						"頑張ってるやん！", "もうちょっと詰めてみようよ！", "君がそんなに考えてわかんないなら，誰もわかんないよ！", 
						"今，頭が回らないだけで少し時間を空けて考えたらわかる時もあるよ！", "それはもう心が１回休めって言ってるんだよ"}

	var docTalk []db_entity.Talk
	for idx, talk := range response {
		docTalk = append(docTalk, db_entity.Talk{ Id: uint(idx), Response: talk })
	}
	toneId := primitive.NewObjectID()
	docTone := &db_entity.BearTone{
		ToneId: toneId,
		ToneName: "test",
		Detail: "テスト",
		Talk: docTalk,
	}

	_, err2 := bearToneCollection.InsertOne(context.TODO(), docTone) // ここでMarshalBSON()される
	if err2 != nil {
		fmt.Println("Error inserting bear")
        panic(err2)
    } else {
		fmt.Println("Successfully inserting bear_tone")
	}
	
	// communityCollection init
	CommunityCollection := db.MongoClient.Database("insertDB").Collection("communities")
	var user_id_array []primitive.ObjectID
	communityId	:= primitive.NewObjectID()
	docCom := &db_entity.Community{
		CommunityId: communityId,
		CommunityName: "test",
		Member: user_id_array,
	}

	_, err3 := CommunityCollection.InsertOne(context.TODO(), docCom) // ここでMarshalBSON()される
	if err3 != nil {
		fmt.Println("Error inserting bear")
        panic(err3)
    } else {
		fmt.Println("Successfully inserting community")
	}

	// userCollection init
	userCollection := db.MongoClient.Database("insertDB").Collection("users")
	var question_id_array []primitive.ObjectID
	var like_id_array []primitive.ObjectID
	docUser := &db_entity.User{
		UserId: primitive.NewObjectID(),
		UserName: "test",
		EmailAddress: "test@example.com",
		Password: "password",
		Icon: "img_dir/test.png",
		Profile: "test",
		CommunityId: communityId,
		Status: "スッキリ",
		Role: db_entity.Role{RoleName: "admin", Permission: 7},
		BearIcon: bearId,
		BearTone: toneId,
		Question: question_id_array,
		Like: like_id_array,
	}

	_, err4 := userCollection.InsertOne(context.TODO(), docUser) // ここでMarshalBSON()される
	if err4 != nil {
		fmt.Println("Error inserting bear")
        panic(err4)
    } else {
		fmt.Println("Successfully inserting users")
	}

	user_id_array = append(user_id_array, docUser.UserId)
	result, err5 := CommunityCollection.UpdateOne(
		context.TODO(),
		bson.M{"_id": communityId},
		bson.D{
			{"$set", bson.D{{"member", user_id_array}, {"updatedAt", time.Now()}}},
		},
	)
	if err5 != nil {
		panic(err5)
	} else {
		fmt.Printf("Updated %v Documents!\n", result.ModifiedCount)
	}
}
