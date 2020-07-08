package models

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"main/core"
)

const DEADLINE_DUPLICATES = 16 * 60 * 60

type Post struct {
	//ID    				primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	RequestId           string `bson:"requestId"`
	Date                string `bson:"date"`
	DateFrom            string `bson:"dateFrom"`
	DateTo              string `bson:"dateTo"`
	DetailsPageUrl      string `bson:"detailsPageUrl"`
	SourceDistrict      string `bson:"sourceDistrict"`
	SourceCity          string `bson:"sourceCity"`
	DestinationDistrict string `bson:"destinationDistrict"`
	DestinationCity    string `bson:"destinationCity"`
	Distance           string `bson:"distance"`
	Truck              string `bson:"truck"`
	SizeMass           string `bson:"sizeMass"`
	SizeMassFrom       string `bson:"sizeMassFrom"`
	SizeMassTo         string `bson:"sizeMassTo"`
	SizeVolume         string `bson:"sizeVolume"`
	SizeVolumeFrom     string `bson:"sizeVolumeFrom"`
	SizeVolumeTo       string `bson:"sizeVolumeTo"`
	Price              string `bson:"price"`
	ProductType        string `bson:"productType"`
	ProductDescription string `bson:"productDescription"`
	ProductComment     string `bson:"productComment"`
	PaymentPrice       string `bson:"productPrice"`
	Dateup             int64  `bson:"dateup"`
}

func (p *Post) GetCountDuplicates() int64 {
	filter := bson.M{ "$or": []bson.M{
		bson.M{
			"requestId": p.RequestId,
		},
		bson.M{
			"sizeMass": p.SizeMass,
			"sourceCity": p.SourceCity,
			"destinationCity": p.DestinationCity,
			"dateup": bson.M{"$gt": p.Dateup - DEADLINE_DUPLICATES},
		},
	}}
	collection := core.GetConnectionMongo().Database("cargodb").Collection("posts")
	count, _ := collection.CountDocuments(context.TODO(), filter)

	return count
}

func (p *Post) Save()  {
	collection := core.GetConnectionMongo().Database("cargodb").Collection("posts")
	_, err := collection.InsertOne(context.TODO(), p)
	if err != nil {
		log.Panic(err)
	}

	//id := res.InsertedID.(primitive.ObjectID)
}

func GetPostByRequestId(requestId string) Post {
	var post Post
	filter := bson.D{{"requestId", requestId}}
	collection := core.GetConnectionMongo().Database("cargodb").Collection("posts")
	err := collection.FindOne(context.TODO(), filter).Decode(&post)
	if err != nil {
		log.Fatal(err)
	}
	return post
}