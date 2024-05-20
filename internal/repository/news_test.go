package repository

import (
	"context"
	"fmt"
	"github.com/chapsuk/grace"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.sport-news/internal/database/mocks"
	"go.sport-news/internal/entity"
	"reflect"
	"testing"
	"time"
)

func setup(t *testing.T) (*Repository, *mocks.DB, context.Context) {
	db := mocks.NewDB(t)
	rep := NewNewsRepository(db)
	ctx := grace.ShutdownContext(context.Background())
	return rep, db, ctx
}

func TestNewNewsRepository(t *testing.T) {
	t.Run("NewNewsRepository created", func(t *testing.T) {
		rep, _, _ := setup(t)

		if !assert.NotNil(t, rep) {
			t.Errorf("NewNewsRepository() = %v, want %s", rep, "not nill")
		}
	})
}

func TestRepository_Exist(t *testing.T) {

	success := []string{"TeamID", "foo"}

	tests := []struct {
		name    string
		want    bool
		wantErr bool
	}{
		{
			name:    "Test exist",
			want:    true,
			wantErr: false,
		},
		{
			name:    "Test don't exist",
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, d, ctx := setup(t)

			var a entity.Article
			c := d.On(
				"FindOne",
				ctx,
				bson.D{
					{Key: success[0], Value: success[1]},
				},
				nil,
				&a,
			)

			if tt.wantErr {
				c.Return(nil, fmt.Errorf("not found"))
			} else {
				c.Return(entity.Article{
					ID:          "",
					TeamID:      success[1],
					ExternalId:  0,
					OptaMatchID: nil,
					Title:       "",
					Type:        nil,
					Teaser:      "",
					Content:     "",
					URL:         "",
					ImageURL:    "",
					GalleryUrls: nil,
					VideoURL:    nil,
					Published:   time.Time{},
				}, nil)
			}

			got, err := r.Exist(ctx, success[0], success[1])
			if (err != nil) != tt.wantErr {
				t.Errorf("Exist() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s() got = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestRepository_GetAllExternalIds(t *testing.T) {

	tests := []struct {
		name string
		want map[int]int
		flag bool
	}{
		{
			name: "Test return extends ids",
			want: map[int]int{1: 1, 2: 2, 3: 3, 4: 4},
			flag: true,
		},
		{
			name: "Test return empty ids",
			want: map[int]int{},
			flag: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, d, ctx := setup(t)

			type re = []struct {
				ExternalId int `bson:"externalId"`
			}
			var resp re
			data := d.On(
				"Find",
				ctx,
				bson.M{},
				&options.FindOptions{Projection: bson.D{{Key: "externalId", Value: 1}}},
				&resp,
			)
			if tt.flag {
				data.Return(re{{ExternalId: 1}, {ExternalId: 2}, {ExternalId: 3}, {ExternalId: 4}}, nil)
			} else {
				data.Return(re{}, nil)
			}

			got, _ := r.GetAllExternalIds(ctx)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAllExternalIds() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepository_GetTeamNews(t *testing.T) {
	articles := []entity.Article{
		{
			ID:          "1",
			TeamID:      "t93",
			ExternalId:  0,
			OptaMatchID: nil,
			Title:       "",
			Type:        nil,
			Teaser:      "",
			Content:     "",
			URL:         "",
			ImageURL:    "",
			GalleryUrls: nil,
			VideoURL:    nil,
			Published:   time.Time{},
		},
		{
			ID:          "2",
			TeamID:      "t94",
			ExternalId:  0,
			OptaMatchID: nil,
			Title:       "",
			Type:        nil,
			Teaser:      "",
			Content:     "",
			URL:         "",
			ImageURL:    "",
			GalleryUrls: nil,
			VideoURL:    nil,
			Published:   time.Time{},
		},
	}
	count := int64(50)
	var a []entity.Article

	tests := []struct {
		name  string
		team  string
		index int
		want  *[]entity.Article
	}{
		{
			name:  "Get teams",
			team:  "t94",
			index: 0,
			want:  &[]entity.Article{articles[0]},
		},
		{
			name:  "Get teams",
			team:  "t93",
			index: 1,
			want:  &[]entity.Article{articles[1]},
		},
		{
			name:  "Get teams",
			team:  "t934",
			index: -1,
			want:  &[]entity.Article{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, d, ctx := setup(t)

			c := d.On(
				"Find",
				ctx,
				bson.D{
					{Key: "teamId", Value: tt.team},
				},
				&options.FindOptions{
					Limit: &count,
					Sort: bson.D{
						{Key: "published", Value: -1},
					},
				},
				&a,
			)
			if tt.index != -1 {
				c.Return(&[]entity.Article{articles[tt.index]}, nil)
			} else {
				c.Return(&[]entity.Article{}, nil)
			}
			got, err := r.GetTeamNews(ctx, tt.team)
			if err != nil {
				t.Errorf("GetTeamNews() error = %v", err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTeamNews() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepository_GetTeamNewsByID(t *testing.T) {
	type args struct {
		team string
		id   string
	}
	e := entity.Article{
		ID:          uuid.New().String(),
		TeamID:      "t93",
		ExternalId:  0,
		OptaMatchID: nil,
		Title:       "",
		Type:        nil,
		Teaser:      "",
		Content:     "",
		URL:         "",
		ImageURL:    "",
		GalleryUrls: nil,
		VideoURL:    nil,
		Published:   time.Time{},
	}

	tests := []struct {
		name    string
		args    args
		want    *entity.Article
		wantErr bool
	}{
		{
			name: "find entity",
			args: args{
				team: e.TeamID,
				id:   e.ID,
			},
			wantErr: false,
			want:    &e,
		},
		{
			name: "not found entity by team",
			args: args{
				team: "",
				id:   e.ID,
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "not found entity by id",
			args: args{
				team: e.TeamID,
				id:   "",
			},
			wantErr: true,
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, d, ctx := setup(t)
			var a entity.Article

			c := d.On(
				"FindOne",
				ctx,
				bson.D{
					{Key: "teamId", Value: tt.args.team},
					{Key: "id", Value: tt.args.id},
				},
				nil,
				&a,
			)
			if tt.args.team == "" {
				c.Return(nil, fmt.Errorf("not found"))
			}
			if tt.args.id == "" {
				c.Return(nil, fmt.Errorf("not found"))
			}
			if tt.args.team != "" && tt.args.id != "" {
				c.Return(&e, nil)
			}

			got, err := r.GetTeamNewsByID(ctx, tt.args.team, tt.args.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetTeamNewsByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTeamNewsByID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepository_InsertArticles(t *testing.T) {

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "check success insert",
			wantErr: false,
		},
		{
			name:    "check error insert",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, d, ctx := setup(t)
			a := []entity.Article{
				{
					ID:          "",
					TeamID:      "",
					ExternalId:  0,
					OptaMatchID: nil,
					Title:       "",
					Type:        nil,
					Teaser:      "",
					Content:     "",
					URL:         "",
					ImageURL:    "",
					GalleryUrls: nil,
					VideoURL:    nil,
					Published:   time.Time{},
				},
			}
			var i []interface{}
			i = append(i, a[0])

			c := d.On("InsertMany", ctx, i)
			if tt.wantErr {
				c.Return(fmt.Errorf("some error"))
			} else {
				c.Return(nil)
			}

			if err := r.InsertArticles(ctx, a); (err != nil) != tt.wantErr {
				t.Errorf("InsertArticles() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRepository_DeleteAll(t *testing.T) {
	tests := []struct {
		name    string
		want    int64
		wantErr bool
	}{
		{name: "test delete All rows", want: 1, wantErr: false},
		{name: "test delete error", want: 0, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, d, ctx := setup(t)

			c := d.On(
				"DeleteMany",
				ctx,
				bson.D{},
				nil,
			)

			if tt.wantErr {
				c.Return(int64(0), fmt.Errorf("some error"))
			} else {
				c.Return(int64(1), nil)
			}

			got, err := r.DeleteAll(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteAll() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeleteAll() got = %v, want %v", got, tt.want)
			}
		})
	}
}
