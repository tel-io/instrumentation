package main

import (
	"context"
	plugin "github.com/tel-io/instrumentation/plugins/mongo"
	"github.com/tel-io/tel/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ccx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		cn := make(chan os.Signal, 1)
		signal.Notify(cn, os.Kill, syscall.SIGINT, syscall.SIGTERM)
		<-cn
		cancel()
	}()

	// Init tel
	cfg := tel.GetConfigFromEnv()
	cfg.LogEncode = "console"
	cfg.Namespace = "TEST"
	cfg.Service = "DEMO-MONGO"
	cfg.LogLevel = "debug"

	t, cc := tel.New(ccx, cfg)
	defer cc()

	// connect to MongoDB
	opts := options.Client()

	// inject plugin
	plugin.Inject(opts, plugin.WithTel(&t))

	opts.ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		panic(err)
	}
	db := client.Database("example")
	inventory := db.Collection("inventory")

	t.Info("Start")

	for i := 0; i < 100; i++ {
		span, ctx := t.StartSpan(ccx, "START_DEMO")
		_, err = inventory.InsertOne(ctx, bson.D{
			{Key: "item", Value: "canvas"},
			{Key: "qty", Value: 100},
			{Key: "attributes", Value: bson.A{"cotton"}},
			{Key: "size", Value: bson.D{
				{Key: "h", Value: 28},
				{Key: "w", Value: 35.5},
				{Key: "uom", Value: "cm"},
			}},
		})
		if err != nil {
			panic(err)
		}

		span.End()
		time.Sleep(time.Second)

	}

	t.Info("End")
}
