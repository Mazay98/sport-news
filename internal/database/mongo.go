package database

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.sport-news/internal/config"
	"go.uber.org/zap"
)

type Mongo struct {
	Client *mongo.Client
	cfg    config.Mongo
}

const (
	articles string = "articles"
)

// MustLoad return new database without errors.
func MustLoad(ctx context.Context, logger *zap.Logger, cfg config.Mongo) DB {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.URL))
	if err != nil {
		logger.Fatal("failed to connect mongo", zap.Error(err))
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		logger.Fatal("failed to ping", zap.Error(err))
	}

	return &Mongo{
		Client: client,
		cfg:    cfg,
	}
}

// Disconnect disconnect from db.
func (m *Mongo) Disconnect(ctx context.Context) {
	if err := m.Client.Disconnect(ctx); err != nil {
		panic(err)
	}
}

// InsertMany insert rows to db.
func (m *Mongo) InsertMany(ctx context.Context, documents []interface{}) error {
	_, err := m.Client.Database(m.cfg.Collection).Collection(articles).InsertMany(ctx, documents)
	return err
}

// Find many rows in db
// filter bson.D{} param
// opts *options.FindOptions
// dataType struct want to be struct.
func (m *Mongo) Find(ctx context.Context, filter interface{}, opts interface{}, dataType interface{}) (interface{}, error) {
	var opt *options.FindOptions
	if opts != nil {
		opt = opts.(*options.FindOptions)
	}

	c, err := m.Client.Database(m.cfg.Collection).Collection(articles).Find(ctx, filter, opt)
	if err != nil {
		return nil, err
	}
	if err = c.All(ctx, dataType); err != nil {
		return nil, err
	}

	return dataType, nil
}

// FindOne  find one  rows in db.
// filter bson.D{} param
// opts *options.FindOptions
// dataType struct want to be struct.
func (m *Mongo) FindOne(ctx context.Context, filter interface{}, opts interface{}, dataType interface{}) (interface{}, error) {
	var opt *options.FindOneOptions
	if opts != nil {
		opt = opts.(*options.FindOneOptions)
	}

	c := m.Client.Database(m.cfg.Collection).Collection(articles).FindOne(ctx, filter, opt)
	if c.Err() != nil {
		return nil, c.Err()
	}

	if err := c.Decode(dataType); err != nil {
		return nil, err
	}

	return dataType, nil
}

// DeleteMany delete many rows.
// filter bson.D{} param
// opts *options.DeleteOptions
func (m *Mongo) DeleteMany(ctx context.Context, filter interface{}, opts interface{}) (int64, error) {
	var opt *options.DeleteOptions
	if opts != nil {
		opt = opts.(*options.DeleteOptions)
	}

	c, err := m.Client.Database(m.cfg.Collection).Collection(articles).DeleteMany(ctx, filter, opt)
	if err != nil {
		return 0, err
	}

	return c.DeletedCount, err
}
