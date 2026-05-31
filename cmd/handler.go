package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"personal/wassup/bl"
	"personal/wassup/config"
	"personal/wassup/db"
	"personal/wassup/db/indexes"
	"personal/wassup/proto"
	"personal/wassup/redis"
	"personal/wassup/transport"
	"personal/wassup/ws"
)

func Init() {
	ctx := context.Background()

	config := config.NewConfig()
	err := config.Init()
	if err != nil {
		panic(err)
	}
	dbHandler := db.NewDBHandler(config)
	err = dbHandler.Init()
	if err != nil {
		panic(err)
	}

	client, err := dbHandler.GetClient()
	if err != nil {
		panic(err)
	}

	indexHandler := indexes.NewIndexHandler(client)
	err = indexHandler.AddIndexes(ctx, db.DBName)
	if err != nil {
		panic(err)
	}

	redisHandler := redis.NewCacheHandler(config)
	err = redisHandler.Init(ctx)
	if err != nil {
		panic(err)
	}

	// psHandler := redis.NewPubSubHandler()
	// err = psHandler.Init(ctx)
	// if err != nil {
	// 	panic(err)
	// }

	webSocketHandler := ws.NewWebSocketHandler(dbHandler, redisHandler)
	wassUpHandler := bl.NewWassupHandler(dbHandler, redisHandler)
	httpHandler := transport.NewHandler(wassUpHandler)
	chiRouter := transport.Router(httpHandler, webSocketHandler)

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	fmt.Println(dir)

	go func() {
		proto.StartGrpcServer(webSocketHandler, redisHandler)
	}()

	if err := http.ListenAndServeTLS(":8080", "./certs/cert.pem", "./certs/key.pem", chiRouter); err != nil {
		panic(err)
	}
}
