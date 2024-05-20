package repository

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.sport-news/internal/database"
	"go.sport-news/internal/entity"
)

//go:generate mockery --name NewsRepository
type NewsRepository interface {
	Exist(ctx context.Context, field string, value any) (bool, error)
	GetTeamNews(ctx context.Context, team string) (*[]entity.Article, error)
	GetTeamNewsByID(ctx context.Context, team, id string) (*entity.Article, error)
	GetAllExternalIds(ctx context.Context) (map[int]int, error)
	InsertArticles(ctx context.Context, articles []entity.Article) error
	DeleteAll(ctx context.Context) (int64, error)
}

type Repository struct {
	db database.DB
}

func NewNewsRepository(db database.DB) *Repository {
	return &Repository{db}
}

// GetTeamNews get articles by team.
func (r *Repository) GetTeamNews(ctx context.Context, team string) (*[]entity.Article, error) {
	var a []entity.Article
	count := int64(50)

	data, err := r.db.Find(
		ctx,
		bson.D{
			{Key: "teamId", Value: team},
		},
		&options.FindOptions{
			Limit: &count,
			Sort: bson.D{
				{Key: "published", Value: -1},
			},
		},
		&a,
	)
	if err != nil {
		return nil, err
	}

	return data.(*[]entity.Article), err
}

// GetTeamNewsByID get article by team and id.
func (r *Repository) GetTeamNewsByID(ctx context.Context, team, id string) (*entity.Article, error) {
	var a entity.Article

	article, err := r.db.FindOne(
		ctx,
		bson.D{
			{Key: "teamId", Value: team},
			{Key: "id", Value: id},
		},
		nil,
		&a,
	)
	if err != nil {
		return nil, err
	}

	return article.(*entity.Article), err
}

// GetAllExternalIds get all ids for match data.
func (r *Repository) GetAllExternalIds(ctx context.Context) (map[int]int, error) {
	type resp = []struct {
		ExternalId int `bson:"externalId"`
	}
	var d resp

	data, err := r.db.Find(ctx, bson.M{}, &options.FindOptions{Projection: bson.D{{Key: "externalId", Value: 1}}}, &d)
	if err != nil {
		return nil, err
	}

	var oldIds map[int]int
	switch v := data.(type) {
	case resp:
		oldIds = make(map[int]int, len(v))
		for _, item := range data.(resp) {
			oldIds[item.ExternalId] = item.ExternalId
		}
	default:
		return make(map[int]int), nil
	}

	return oldIds, nil
}

// InsertArticles insert many articles to db.
func (r *Repository) InsertArticles(ctx context.Context, articles []entity.Article) error {
	var interfaceSlice []interface{} // todo: мб можно убрать костыль
	for _, a := range articles {
		interfaceSlice = append(interfaceSlice, a)
	}

	return r.db.InsertMany(ctx, interfaceSlice)
}

// Exist check for exist field with value on db.
func (r *Repository) Exist(ctx context.Context, field string, value any) (bool, error) {
	var a entity.Article
	_, err := r.db.FindOne(
		ctx,
		bson.D{
			{Key: field, Value: value},
		},
		nil,
		&a,
	)
	if err != nil {
		return false, err
	}
	return true, nil
}

// DeleteAll delete all rows.
func (r *Repository) DeleteAll(ctx context.Context) (int64, error) {
	return r.db.DeleteMany(ctx, bson.D{}, nil)
}
