package main

import (
	"database/sql"
	"fmt"
	"os"

	"server/kafka"
	"server/models"
	_ "server/routers"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/filter/cors"
	"github.com/beego/beego/v2/core/logs"
	_ "github.com/go-sql-driver/mysql"
)

func getConfigString(key string) string {
	value := beego.AppConfig.DefaultString(key, "")
	if value == "" {
		logs.Error("Failed to get config value for %s", key)
		os.Exit(1)
	}
	return value
}

func init() {
	// Load configuration
	err := beego.LoadAppConfig("ini", "conf/app.conf")
	if err != nil {
		logs.Error("Failed to load configuration: %v", err)
		os.Exit(1)
	}

	// Initialize Kafka
	kafkaBrokers := []string{getConfigString("kafka_brokers")}
	err = kafka.InitKafka(kafkaBrokers)
	if err != nil {
		logs.Error("Failed to initialize Kafka: %v", err)
		os.Exit(1)
	}

	// Initialize Kafka consumers
	err = models.InitializeKafkaConsumers()
	if err != nil {
		logs.Error("Failed to initialize Kafka consumers: %v", err)
		os.Exit(1)
	}

	// MySQL 연결 설정
	dbUser := beego.AppConfig.DefaultString("db_user", "")
	dbPass := beego.AppConfig.DefaultString("db_pass", "")
	dbName := beego.AppConfig.DefaultString("db_name", "")
	dbHost := beego.AppConfig.DefaultString("db_host", "localhost")
	dbPort := beego.AppConfig.DefaultString("db_port", "3306")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		logs.Error("Failed to connect to MySQL: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		logs.Error("Failed to ping MySQL: %v", err)
		os.Exit(1)
	}

	logs.Info("Successfully connected to MySQL")
}

func main() {
	// CORS 설정
	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Access-Control-Allow-Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin"},
		AllowCredentials: true,
	}))

	// WebSocket 설정
	beego.SetStaticPath("/ws", "websocket")

	// 로깅 설정
	logConfig := `{
		"filename": "logs/project.log",
		"level": 7,
		"maxlines": 1000000,
		"maxsize": 268435456,
		"daily": true,
		"maxdays": 7
	}`
	err := logs.SetLogger(logs.AdapterFile, logConfig)
	if err != nil {
		logs.Error("Failed to set logger: %v", err)
		os.Exit(1)
	}
	logs.SetLevel(logs.LevelDebug)
	logs.EnableFuncCallDepth(true)
	logs.Async()

	// Beego 애플리케이션 실행
	defer kafka.CloseKafka()
	beego.Run()
}

