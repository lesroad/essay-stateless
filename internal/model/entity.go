package model

import (
	"time"
	
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RawLogs struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	URL        string             `bson:"url"`
	Request    string             `bson:"request"`
	Response   string             `bson:"response"`
	CreateTime time.Time          `bson:"create_time"`
} 