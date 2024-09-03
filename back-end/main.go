package main

//这个后端的http Server部分用了GIN框架，因为GIN提供的接口比较简单
//数据库部分用了gorm里面的SQLite的driver
import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

// Comment 是评论的数据库模型，因为Go的导出机制，这里的字段必须首字母大写，不然导不出去
// 这也就说明了序列化时为了严格遵守接口文档必须要加上json标签来指定这些字段全部小写
type Comment struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	Name    string `gorm:"size:255"   json:"name"`
	Content string `gorm:"type:text"  json:"content"`
}

// Response data是空接口类型，即any,根据接口文档做序列化即可
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// 初始化数据库，这里启动不了数据库就直接panic，不然服务器启动了也没用还得重启
func initDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("comments.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	return db
}

var db *gorm.DB

func main() {
	db = initDB()
	// 启动一个GIN
	r := gin.Default()

	// 配置 CORS 中间件，如果没有配置，add的时候会抛出CORS error，应该是前后端不在一个域的原因
	// 注意到浏览器在发add请求前还发了一个preflight请求，这个请求的方法是OPTIONS，相当于检查服务器有没有设置CORS功能
	// 直接用GIN的中间件配置cors即可，Default方法的配置是：
	//access-control-allow-headers: Origin,Content-Length,Content-Type
	//access-control-allow-methods: GET,POST,PUT,PATCH,DELETE,HEAD,OPTIONS
	//access-control-allow-origin: *
	//access-control-max-age: 43200
	r.Use(cors.Default())

	// 路由，分别对应增删查的request，跳到对应的handler处理
	r.GET("/comment/get", getComments)
	r.POST("/comment/add", addComment)
	r.POST("/comment/delete", deleteComment)
	// 启动服务器，监听在8080端口(虽然默认也是8080)
	r.Run(":8080")
}

// 获取评论
// gin.Context类型封装了http.request和http.ResponseWriter，以及一系列方法用来查找和收发数据，比原来的http包更好用
func getComments(c *gin.Context) {
	var comments []Comment
	var total int64
	// 用Query方法查找url里面的查询参数，然后转成int
	pageStr := c.Query("page")
	sizeStr := c.Query("size")

	page, _ := strconv.Atoi(pageStr)
	size, _ := strconv.Atoi(sizeStr)

	// 数据库操作，如果size等于-1就找所有的comments，不等于就根据size和page的分页规则找
	// 这里用了gorm里面的Find Model Offset Limit Count方法
	// Model方法用来根据传入参数的类型去匹配数据库里的一个表，相当于定位
	// Count把这个表里的元组总数传到传入的参数里
	// Find返回表里的数据
	// Offset和Limit用来实现分页，Offset就是从表的哪里开始，Limit就是指定Find的数量
	if size == -1 {
		db.Find(&comments)
		db.Model(&Comment{}).Count(&total)
	} else {
		db.Model(&Comment{}).Count(&total)
		db.Offset((page - 1) * size).Limit(size).Find(&comments)
	}
	// gin.Context的JSON方法用来发送JSON的数据，data直接用string到any的匿名映射发
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "success",
		Data: map[string]interface{}{
			"total":    total,
			"comments": comments,
		},
	})
}

// 添加评论
func addComment(c *gin.Context) {
	var input Comment
	// ShouldBindJSON方法用来反序列化发来的数据，把它解析到input结构体的各个字段里
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 1, Msg: "Invalid input"})
		return
	}

	// 创建评论
	db.Create(&input)
	// 发成功的response
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "success",
		Data: input,
	})
}

// 删除评论
func deleteComment(c *gin.Context) {
	// 找请求里的id,如果不是整数，就报错，然后做对应的错误处理
	// 因为前端的删除按钮是和id绑定的，不存在传入的数字不匹配评论的情况，只存在转换过程出错的情况，所以只需要判断是否是整数就行
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 1, Msg: "Invalid ID"})
		return
	}

	db.Delete(&Comment{}, id)
	// 发成功的response
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "success",
		Data: nil,
	})
}
