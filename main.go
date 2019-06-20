package main

import (
	"fmt"
	"log"
	"net/http"

	"database/sql"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

const (
	redisAddr     = "127.0.0.1:6379"
	redisPassword = ""
	redisDB       = 1
	sidKey        = "else.sid"

	mysqldns = "root:ipassword@tcp(localhost:3306)/hello"
)

var (
	redisClient *redis.Client
	mysqlDB     *sql.DB
)

// 完成redis与mysql连接的初始化
func init() {
	redisClient = defaultRedisClient()
	mysqlDB = defaultMysqlDB()
}

// ------------------
// ---- server ------
// ------------------

func main() {
	defer closeDefaultMysqlDB()
	// 实例化一个Echo类型
	e := echo.New()
	// 添加中间件
	e.Use(middleware.Logger())
	// 添加路由
	e.GET("/helloredis", helloredis, middlewareValidateSession())
	e.GET("/hellomysql", hellomysql)
	// 启动服务器
	e.Logger.Fatal(e.Start(":1323"))
}

// --------------------
// ----- database------
// --------------------

func defaultRedisClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})
	pong, err := client.Ping().Result()
	log.Println(pong, err)
	return client
}

func defaultMysqlDB() *sql.DB {
	db, err := sql.Open("mysql", mysqldns)
	if err != nil {
		log.Println("err: sql.Open('sql', %q), %s", mysqldns, err)
	}

	err = db.Ping()
	if err != nil {
		log.Println("err: db.Ping(), %s", err)
	}
	log.Printf("mysqlDB connected...")
	return db
}

func closeDefaultMysqlDB() {
	mysqlDB.Close()
}

// --------------------
// ----- handler ------
// --------------------

func helloredis(c echo.Context) error {
	return c.JSON(http.StatusOK, "Hello, redis")
}

func hellomysql(c echo.Context) error {
	var (
		name, salary string
		age          int
	)
	row := mysqlDB.QueryRow("select name, salary, age from employees limit 1")
	err := row.Scan(&name, &salary, &age)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, fmt.Sprintf("Hello mysql: row(name: %s, salary: %s, age: %d)", name, salary, age))
}

// --------------------
// ---- middleware ----
// --------------------

// 中间件: 进行session验证
func middlewareValidateSession() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 尝试获取session id
			sidCookie, err := c.Cookie(sidKey)
			if err != nil {
				return c.JSON(http.StatusForbidden, "no session id in cookie !")
			}
			// 尝试获取session
			session, err := redisClient.Do("Get", sidCookie.Value).String()
			if err != nil {
				log.Printf("err: redisClient.Do(%v) session=%q err=%v \n", sidCookie.Value, session, err)
				return c.JSON(http.StatusForbidden, "session verification failed !")
			}
			return next(c)
		}
	}
}
