package core

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongodb *mongo.Client

func DisconnectMongo() {
	_ = mongodb.Disconnect(context.TODO())
	mongodb = nil
}

func GetConnectionMongo() *mongo.Client {
	if mongodb == nil {
		mongodb, _ = mongo.Connect(context.TODO(), options.Client().ApplyURI(Config.MongodbUri))
	}

	return mongodb
}
