package user_controller

import (
	"back-challe-chara2022/entity/request_entity/body"
	"back-challe-chara2022/db"
	"back-challe-chara2022/s3"
	// "back-challe-chara2022/entity/db_entity"

	"net/http"
	"fmt"
	"context"
	"strings"
	
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"github.com/gin-gonic/gin"
	jwt "github.com/appleboy/gin-jwt/v2"
)

type UserController struct {}

type UserResponse struct {
	UserName string `json:"userName"`
	Profile string `json:"profile"`
	Status string `json:"status"`
}

type UserIconResponse struct {
	UserIcon []byte `json:"userIcon"`
}

type UserCommunityResponse struct {
	CommunityId primitive.ObjectID `json:"communityId"`
	CommunityName string `json:"communityName"`
	Icon []byte `json:"icon"`
}

type UserStatusResponse struct {
	IsUpdated bool `json:"isUpdated"`
}

type DocCommunity struct {
	Id primitive.ObjectID `json:"id"`
	CommunityId []string `json:"communityId"`
}

// GET: /user
func (uc UserController) GetUser(c *gin.Context) {

	// ユーザ情報を返すAPI
	
	var err error

	claims := jwt.ExtractClaims(c)
	userId, _ := primitive.ObjectIDFromHex(claims["userId"].(string))
	fmt.Println(userId) // debug message

	userCollection := db.MongoClient.Database("insertDB").Collection("users")
	var doc bson.M
	// 検索条件
	filter := bson.D{{"_id", userId}}
	// query the user collection
	err = userCollection.FindOne(context.TODO(), filter,
		options.FindOne().SetProjection(bson.M{"userName": 1, "profile": 1, "status": 1, "_id": 0})).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		fmt.Printf("No document was found with the stampId")
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"message": "No document was found with the stampId",
		})
		return
	} else if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"message": err.Error(),
		})
		return
	}

	// Response
	c.JSON(http.StatusOK, UserResponse{
		UserName: doc["userName"].(string),
		Profile: doc["profile"].(string),
		Status: doc["status"].(string),
	})
	return
}

// PATCH: /user/status
func (uc UserController) PatchUserStatus(c *gin.Context) {

	// スタンプが押された際に，userのステータスを更新

	claims := jwt.ExtractClaims(c)
	userId, _ := primitive.ObjectIDFromHex(claims["userId"].(string))
	fmt.Println(userId) // debug message

	var request body.PatchUserStatusBody
	// bodyのjsonデータを構造体にBind
	if err := c.Bind(&request); err != nil {
		// bodyのjson形式が合っていない場合
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
		return
	}
	fmt.Println(request.StampId) // debug print

	var docStamp bson.M
	// 検索条件
	stampId, _ := primitive.ObjectIDFromHex(request.StampId)
	filterStamp := bson.D{{"_id", stampId}}
	fmt.Println(filterStamp)
	// query to stampCollection
	stampCollection := db.MongoClient.Database("insertDB").Collection("stamps")
	if err := stampCollection.FindOne(context.TODO(), filterStamp, 
		options.FindOne().SetProjection(bson.M{"status": 1, "_id": 0})).Decode(&docStamp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"message": err.Error(),
		})
	} else if err == mongo.ErrNoDocuments {
		fmt.Printf("No document was found with the stampId")
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"message": "No document was found with the stampId",
		})
		return
	}

	fmt.Println(docStamp)
	// update raw data
	updateFields := bson.M{
		"$set": bson.M{
			"status": docStamp["status"].(string),
		},
	}
	filter := bson.M{"_id": userId}
	userCollection := db.MongoClient.Database("insertDB").Collection("users")
	result, err := userCollection.UpdateOne(context.TODO(), filter, updateFields)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
		return
	} else if err == mongo.ErrNoDocuments {
		fmt.Printf("No document was found with the userId")
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"message": "No document was found with the userId",
		})
		return
	}

	fmt.Println(result)

	var response UserStatusResponse

	if result.ModifiedCount > 0 {
		response.IsUpdated = true
	} else {
		response.IsUpdated = false
	}

	c.JSON(http.StatusOK, response)
	return
}

// GET: /user/community
// $inを用いることで1つのクエリでいけるかも
func (uc UserController) GetUserCommunity(c *gin.Context) {

	// userIdが所属するコミュニティのcommunity_nameを全て返す

	claims := jwt.ExtractClaims(c)
	userId, _ := primitive.ObjectIDFromHex(claims["userId"].(string))
	fmt.Println(userId) // debug message

	var err error

	userCollection := db.MongoClient.Database("insertDB").Collection("users")

	var doc_filter bson.Raw
	// 検索条件
	filter := bson.D{{"_id", userId}}
	// query the user collection
	err = userCollection.FindOne(context.TODO(), filter, options.FindOne().SetProjection(bson.M{"communityId": 1})).Decode(&doc_filter)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
		return
	} else if err == mongo.ErrNoDocuments {
		fmt.Printf("No document was found with the userId")
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"message": "No document was found with the userId",
		})
		return
	}
	var d_tmp DocCommunity

	// 配列の型を確定させるためにbsonを構造体に変換
	err = bson.Unmarshal(doc_filter, &d_tmp)
	if len(d_tmp.CommunityId) < 1 {
		c.JSON(http.StatusOK, gin.H{"userCommunity": d_tmp.CommunityId})
		return
	}

	var response []UserCommunityResponse

	for _, doc := range d_tmp.CommunityId {

		var docCommunity bson.M
		// 検索条件
		id, _ := primitive.ObjectIDFromHex(doc)
 		filterCommunity := bson.M{"_id": id}
		// query the community collection
		communityCollection := db.MongoClient.Database("insertDB").Collection("communities")
		err = communityCollection.FindOne(context.TODO(), filterCommunity, options.FindOne().SetProjection(bson.M{"communityName": 1, "icon": 1, "_id": 0})).Decode(&docCommunity)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
			return
		} else if err == mongo.ErrNoDocuments {
			fmt.Printf("No document was found with the userId")
			// c.JSON(http.StatusOK, gin.H{"userCommunity": make([]string, 0)})
			c.JSON(http.StatusNotFound, gin.H{
				"code": 404,
				"message": "No document was found with the userId",
			})
			return
		}
		
		// 画像の処理
		var url string = docCommunity["icon"].(string)
		var bucketIndex int = strings.Index(url, "/") // 最初に "/" が出現する位置
		var bucketName, key string = url[:bucketIndex], url[bucketIndex:]
		// S3インスタンスを作成
		s3Instance, err := s3.NewS3()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"code": 503,
				"message": "Service Unavailable",
			})
		}
		// S3から画像ファイルのダウンロード
		downloadKey := s3.GetObjectInput(bucketName, key)
		buf, err := s3.Download(s3Instance, downloadKey) //[]byte
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"code": 404,
				"message": err.Error(),
			})
		}

		// Responseデータに追加
		response = append(response, UserCommunityResponse{
			CommunityId: id,
			CommunityName: docCommunity["communityName"].(string),
			Icon: buf,
		})

	}

	c.JSON(http.StatusOK, gin.H{
		"communities": response,
	})
	return


}

// GET: /user/icon
func (uc UserController) GetUserIcon(c *gin.Context) {

	// userIdのユーザのiconを返す

	claims := jwt.ExtractClaims(c)
	userId, _ := primitive.ObjectIDFromHex(claims["userId"].(string))
	fmt.Println(userId) // debug message

	var err error
	userCollection := db.MongoClient.Database("insertDB").Collection("users")
	var doc bson.M
	// 検索条件
	filter := bson.M{"_id": userId}
	// query
	if err := userCollection.FindOne(context.TODO(), filter, 
		options.FindOne().SetProjection(bson.M{"icon": 1, "_id": 0})).Decode(&doc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
		return
	} else if err == mongo.ErrNoDocuments {
		fmt.Printf("No document was found with the userId")
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"message": "No document was found with the userId",
		})
		return
	}

	// S3バケットとオブジェクトを指定
	url := doc["icon"].(string)
	var bucketIndex int = strings.Index(url, "/") // 最初に "/" が出現する位置
	var bucketName, key string = url[:bucketIndex], url[bucketIndex:]
	
	fmt.Println(bucketName, key)
	
	// S3インスタンスを作成
	s3Instance, err := s3.NewS3()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"code": 503,
			"message": "Service Unavailable",
		})
	}

	// S3から画像ファイルのダウンロード
	downloadKey := s3.GetObjectInput(bucketName, key)
	buf, err := s3.Download(s3Instance, downloadKey) //[]byte
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"code": 404,
			"message": err.Error(),
		})
	} else {
		response := UserIconResponse{UserIcon: buf}
		c.JSON(http.StatusOK, response)
	}
}