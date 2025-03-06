package structurs

import "go.mongodb.org/mongo-driver/bson/primitive"

type Item struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Name       string             `bson:"name"`
	ImageURL   string             `bson:"imageURL"`
	Popularity int                `bson:"popularity"`
}
