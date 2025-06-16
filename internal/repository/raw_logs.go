package repository

import (
	"context"
	"essay-stateless/internal/model"

	"go.mongodb.org/mongo-driver/mongo"
)

type RawLogsRepository interface {
	Save(ctx context.Context, log *model.RawLogs) error
}

type rawLogsRepository struct {
	collection *mongo.Collection
}

func NewRawLogsRepository(db *mongo.Database) RawLogsRepository {
	return &rawLogsRepository{
		collection: db.Collection("sts_logs"),
	}
}

func (r *rawLogsRepository) Save(ctx context.Context, log *model.RawLogs) error {
	_, err := r.collection.InsertOne(ctx, log)
	return err
}
