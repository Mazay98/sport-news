package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.sport-news/internal/config"
	"go.sport-news/internal/database"
	"go.sport-news/internal/entity"
	ll "go.sport-news/internal/logger"
	"go.sport-news/internal/repository"
	"go.sport-news/internal/scheduler"
	"go.uber.org/zap"
	"net/http"
	"testing"
	"time"
)

func Test_e2e(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	logger := ll.MustLoad("", "test", "debug")
	cfg := config.MustLoadFromYAML("e2e.yml")

	db := database.MustLoad(ctx, logger, config.Mongo{
		URL:        cfg.Mongo.URL,
		Collection: cfg.Mongo.Collection,
	})
	defer db.Disconnect(ctx)
	repo := repository.NewNewsRepository(db)

	_, err := repo.DeleteAll(ctx)
	if err != nil {
		t.Error(err)
	}

	err = scheduler.New(logger, config.Parser{
		URL:     fmt.Sprintf("http://0.0.0.0:%d", cfg.HTTP.ExternalPort),
		Count:   1,
		JobTime: time.Minute,
	}, db).RunNow()

	if !assert.Nil(t, err) {
		logger.Fatal("failed run job", zap.Error(err))
		return
	}

	<-time.NewTimer(time.Second).C

	list, err := repo.GetTeamNews(ctx, entity.DefaultTeamId)
	if err != nil {
		t.Error(err)
	}

	assert.Len(t, *list, 1)

	for _, article := range *list {
		_, err := http.Post(fmt.Sprintf("http://0.0.0.0:%d/v1/cache-flush", cfg.HTTP.Port), "application/json", nil)
		if err != nil {
			t.Error(err)
		}

		resp, err := http.Get(fmt.Sprintf("http://0.0.0.0:%d/v1/teams/%s/news", cfg.HTTP.Port, article.TeamID))
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()

		var data struct {
			Status  string
			Data    []entity.Article
			Message string
		}
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			t.Error(err)
		}

		fmt.Println(data)

		assert.Equal(t, "success", data.Status, "except equ statuses")
		assert.Equal(t, article.ID, data.Data[0].ID, "except equ Id")

		resp, err = http.Get(fmt.Sprintf("http://0.0.0.0:%d/v1/teams/%s/news/%s", cfg.HTTP.Port, article.TeamID, article.ID))
		if err != nil {
			t.Error(err)
		}
		var sData struct {
			Status string
			Data   entity.Article
		}
		if err := json.NewDecoder(resp.Body).Decode(&sData); err != nil {
			t.Error(err)
		}

		assert.Equal(t, "success", data.Status, "except equ statuses")
		assert.Equal(t, article.ID, sData.Data.ID, "except equ Id")
	}

}
