package main

import (
	"context"
	"sync"
	"time"

  "go.uber.org/zap"

	redis "github.com/go-redis/redis"
	kafka "github.com/segmentio/kafka-go"
)

var client *redis.Client

var logger *zap.Logger
var sugar *zap.SugaredLogger

func main() {
  logger, _ = zap.NewProduction()
	defer logger.Sync()
	sugar = logger.Sugar()

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"kafka:9092"},
		Topic:   "Articles",
		Partition: 0,
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MBkafka
	})

	client = redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})
	defer client.Close()

	_, err := client.Ping().Result()
	if err != nil {
		sugar.Error(err)
	}

	var wg sync.WaitGroup
	c := 0

	for {

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		msg, err := r.ReadMessage(ctx)
		if err != nil {
			sugar.Info(err)
		}

		wg.Add(1)

		goCtx, goCcancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer goCcancel()

		go process(goCtx, c, &wg, msg)
		c++
	}

	wg.Wait()

	err = r.Close()
	if err != nil {
		sugar.Error(err)
	}
}

func process(ctx context.Context, counter int, wg *sync.WaitGroup, msg kafka.Message) {
	defer wg.Done()

	massage := string(msg.Value)
  if massage == "" {
    return
  }

  ////////////////
  // do somthing
  ////////////////

  err := add_to_db(massage)
  if err != nil {
			sugar.Error(err)
	}
}

func add_to_db(massage string) (err error) {
	cmd := redis.NewStringCmd("select", 0)
	err = client.Process(cmd)
  post_time := time.Now().String()
	err = client.Set(post_time, massage, 0).Err()
	return err
}
