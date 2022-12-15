package server

import (
	"back-challe-chara2022/controller/bear_controller"
	"back-challe-chara2022/controller/user_controller"
	"back-challe-chara2022/controller/login_controller"
	
	"time"
	"os"
	
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
)

// 初期化
func Init() {

	// ルーティング
	r := setRouter()
	// Server Run (Port 8080)
	if err := r.Run(":" + os.Getenv("PORT")); err != nil {
		panic(err)
	}
}

// ルーティング設定
func setRouter() *gin.Engine {
	
	r := gin.Default()

	// ミドルウェアの設定
	r.Use(cors.New(cors.Config{
		// アクセスを許可したいアクセス元
		AllowOrigins: []string{
			"http://localhost", 
			"http://localhost:3000",
			"https://qmatta.vercel.app",
		},
		// アクセスを許可したいHTTPメソッド
		AllowMethods: []string{
			"POST",
			"GET",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},
		// 許可したいHTTPリクエストヘッダ
		AllowHeaders: []string{
			"Access-Control-Allow-Credentials",
			"Access-Control-Allow-Origin",
			"Access-Control-Allow-Headers",
			"Content-Type",
			"Content-Length",
			"Accept-Encoding",
			"Authorization",
		},
		// cookieなどの情報を必要とするかどうか
		AllowCredentials: true,
		// preflightリクエストの結果をキャッシュする時間
		MaxAge: 24 * time.Hour,
	  }))

	//ルーティング
	bear_group := r.Group("bear")
	{
		ctrl := bear_controller.BearController{}
		// 熊の返答を返す
		bear_group.GET("/", ctrl.GetNotLoginResponse) // not required login user 
		bear_group.POST(":userId", ctrl.PostResponse) // required login user
		// クマとの対話履歴を返す
		bear_group.GET("history/:userId", ctrl.GetHistory)
	}

	user_group := r.Group("user")
	{
		ctrl := user_controller.UserController{}
		// userのステータスを更新
		user_group.PATCH("status/:userId", ctrl.PatchUserStatus)
		// userの所属するコミュニティを全て取得
		user_group.GET("community/:userId", ctrl.GetUserCommunity)
		// userのアイコンを取得
		user_group.GET("icon/:userId", ctrl.GetUserIcon)	
	}

	// user登録
	r.POST("/signup", login_controller.CreateUser)
	// ユーザ認証
	r.POST("/login", login_controller.LoginUser)

	return r
}

