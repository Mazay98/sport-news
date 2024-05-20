package database

import "context"

//go:generate go run github.com/vektra/mockery/v2@v2.42.0 --name DB
type DB interface {
	Disconnect(ctx context.Context)
	InsertMany(ctx context.Context, documents []interface{}) error
	Find(ctx context.Context, filter interface{}, opts interface{}, dataType interface{}) (interface{}, error)
	FindOne(ctx context.Context, filter interface{}, opts interface{}, dataType interface{}) (interface{}, error)
	DeleteMany(ctx context.Context, filter interface{}, opts interface{}) (int64, error)
}
