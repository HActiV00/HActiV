package main

import (
	"server/kafka"
	"server/models"
	_ "server/routers"
	"time"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/filter/cors"
	"github.com/beego/beego/v2/core/logs"
)

func getConfigString(key string) string {
	value := beego.AppConfig.DefaultString(key, "")
	if value == "" {
		logs.Error("Failed to get config value for %s", key)
		panic("Missing configuration")
	}
	return value
}

func initKafkaWithRetry(maxRetries int, retryInterval time.Duration) error {
	kafkaBrokers := []string{getConfigString("kafka_brokers")}
	var err error

	for i := 0; i < maxRetries; i++ {
		err = kafka.InitKafka(kafkaBrokers)
		if err == nil {
			logs.Info("Successfully connected to Kafka")
			return nil
		}
		logs.Warn("Failed to connect to Kafka, retrying in %v... (Attempt %d/%d)", retryInterval, i+1, maxRetries)
		time.Sleep(retryInterval)
	}

	return err
}

func init() {
	// Load configuration
	err := beego.LoadAppConfig("ini", "conf/app.conf")
	if err != nil {
		logs.Error("Failed to load configuration: %v", err)
		panic(err)
	}

	// Initialize Kafka with retry
	err = initKafkaWithRetry(5, 5*time.Second)
	if err != nil {
		logs.Error("Failed to initialize Kafka after multiple attempts: %v", err)
		panic(err)
	}

	// Initialize Kafka consumers
	err = models.InitializeKafkaConsumers()
	if err != nil {
		logs.Error("Failed to initialize Kafka consumers: %v", err)
		panic(err)
	}
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
		panic(err)
	}
	logs.SetLevel(logs.LevelDebug)
	logs.EnableFuncCallDepth(true)
	logs.Async()

	// Beego 애플리케이션 실행
	defer kafka.CloseKafka()
	beego.Run()
}
