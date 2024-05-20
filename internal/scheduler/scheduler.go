package scheduler

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	cloudflarebp "github.com/DaRealFreak/cloudflare-bp-go"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"go.sport-news/internal/config"
	"go.sport-news/internal/database"
	"go.sport-news/internal/entity"
	"go.sport-news/internal/repository"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"io"
	"net/http"
	"sync"
	"time"
)

type NewsItem struct {
	ArticleURL        string `xml:"ArticleURL"`
	NewsArticleID     int    `xml:"NewsArticleID"`
	PublishDate       string `xml:"PublishDate"`
	Taxonomies        string `xml:"Taxonomies"`
	TeaserText        string `xml:"TeaserText"`
	ThumbnailImageURL string `xml:"ThumbnailImageURL"`
	Title             string `xml:"Title"`
	OptaMatchId       string `xml:"OptaMatchId"`
	LastUpdateDate    string `xml:"LastUpdateDate"`
	IsPublished       string `xml:"IsPublished"`
}
type News struct {
	XMLName             xml.Name `xml:"NewListInformation"`
	ClubName            string   `xml:"ClubName"`
	ClubWebsiteURL      string   `xml:"ClubWebsiteURL"`
	NewsletterNewsItems struct {
		NewsletterNewsItem []NewsItem `xml:"NewsletterNewsItem"`
	} `xml:"NewsletterNewsItems"`
}
type NewsArticleInformation struct {
	XMLName        xml.Name `xml:"NewsArticleInformation"`
	ClubName       string   `xml:"ClubName"`
	ClubWebsiteURL string   `xml:"ClubWebsiteURL"`
	NewsArticle    struct {
		ArticleURL        string `xml:"ArticleURL"`
		NewsArticleID     int    `xml:"NewsArticleID"`
		PublishDate       string `xml:"PublishDate"`
		Taxonomies        string `xml:"Taxonomies"`
		TeaserText        string `xml:"TeaserText"`
		Subtitle          string `xml:"Subtitle"`
		ThumbnailImageURL string `xml:"ThumbnailImageURL"`
		Title             string `xml:"Title"`
		BodyText          string `xml:"BodyText"`
		GalleryImageURLs  string `xml:"GalleryImageURLs"`
		VideoURL          string `xml:"VideoURL"`
		OptaMatchId       string `xml:"OptaMatchId"`
		LastUpdateDate    string `xml:"LastUpdateDate"`
		IsPublished       string `xml:"IsPublished"`
	} `xml:"NewsArticle"`
}

func New(logger *zap.Logger, cfg config.Parser, db database.DB) gocron.Job {
	s, err := gocron.NewScheduler()
	if err != nil {
		logger.Fatal("failed init scheduler", zap.Error(err))
	}

	j, err := s.NewJob(
		gocron.DurationJob(
			cfg.JobTime,
		),
		gocron.NewTask(
			task,
			logger,
			cfg,
			db,
		),
	)
	if err != nil {
		logger.Fatal("failed register job", zap.Error(err))
	}

	logger.Info("register job", zap.String("uuid", j.ID().String()))
	s.Start()

	return j
}

// task for scheduler
// he goes to http and parse xml files.
func task(logger *zap.Logger, cfg config.Parser, db database.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	rep := repository.NewNewsRepository(db)
	oldIds, err := rep.GetAllExternalIds(ctx)
	if err != nil {
		logger.Error("failed get GetAllExternalIds", zap.Error(err))
		return
	}

	var (
		a   News
		url = fmt.Sprintf("%s/getnewlistinformation?count=%d", cfg.URL, cfg.Count)
	)
	err = makeRequest(ctx, url, &a, 3)
	if err != nil {
		logger.Error("failed do request", zap.Error(err), zap.String("url", url), zap.Any("data", a))
		return
	}

	newIds := make(map[int]NewsItem)
	for _, item := range a.NewsletterNewsItems.NewsletterNewsItem {
		_, ok := oldIds[item.NewsArticleID]
		if !ok {
			newIds[item.NewsArticleID] = item
		}
	}

	wg := sync.WaitGroup{}
	wg.Add(len(newIds))
	mu := sync.Mutex{}
	var addData []entity.Article

	for _, item := range newIds {
		go func(wg *sync.WaitGroup, mu *sync.Mutex, item NewsItem) {
			defer wg.Done()

			t, err := time.Parse(time.DateTime, item.PublishDate)
			if err != nil {
				logger.Error("failed parse date", zap.Error(err), zap.String("data", item.PublishDate))
				return
			}

			var (
				aI  NewsArticleInformation
				url = fmt.Sprintf("%s/getnewsarticleinformation?id=%d", cfg.URL, item.NewsArticleID)
			)
			if err = makeRequest(ctx, url, &aI, 0); err != nil {
				logger.Error("failed do request", zap.Error(err), zap.String("url", url), zap.Any("data", aI))
				return
			}

			mu.Lock()
			addData = append(addData, entity.Article{
				ID:          uuid.New().String(),
				TeamID:      entity.DefaultTeamId,
				ExternalId:  item.NewsArticleID,
				OptaMatchID: &item.OptaMatchId,
				Title:       item.Title,
				Type:        []string{item.Taxonomies},
				Teaser:      item.TeaserText,
				Content:     aI.NewsArticle.BodyText,
				URL:         item.ArticleURL,
				ImageURL:    item.ThumbnailImageURL,
				GalleryUrls: aI.NewsArticle.GalleryImageURLs,
				VideoURL:    aI.NewsArticle.VideoURL,
				Published:   t,
			})
			mu.Unlock()
		}(&wg, &mu, item)
	}
	wg.Wait()

	if len(addData) > 0 {
		err = rep.InsertArticles(ctx, addData)
		if err != nil {
			logger.Error("failed add posts", zap.Error(err))
			return
		}
	}
	logger.Info("success done job", zap.Int("added posts", len(addData)))
}

// makeRequest do a http get request with retry.
func makeRequest(ctx context.Context, url string, data any, retries int) error {
	var (
		response *http.Response
		try      = 0
		te       error
	)
	timeout := time.Second * 20

	for try <= retries {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, bytes.NewBuffer(nil))
		if err != nil {
			te = multierr.Append(te, err)
			try++
			continue
		}

		c := &http.Client{Timeout: timeout}
		c.Transport = cloudflarebp.AddCloudFlareByPass(c.Transport)

		response, err = c.Do(req)

		if response == nil || err != nil {
			try++
			continue
		}

		switch response.StatusCode {
		case http.StatusInternalServerError:
			body, _ := io.ReadAll(response.Body)
			te = multierr.Append(te, fmt.Errorf("%s", fmt.Sprintf("err[500],req:\n%s", body)))
			response.Body.Close() //nolint:errcheck
			time.Sleep(time.Second)
			try++
			continue
		case http.StatusBadRequest:
			body, _ := io.ReadAll(response.Body)
			te = multierr.Append(te, fmt.Errorf("err[400],req:\n%s, data:\n%s", url, body))
			response.Body.Close() //nolint:errcheck
			return te
		}

		break
	}

	if response == nil {
		return multierr.Append(te, fmt.Errorf("failed get response (empty response) in %s", http.MethodGet))
	}
	defer response.Body.Close() //nolint:errcheck

	if err := xml.NewDecoder(response.Body).Decode(&data); err != nil {
		body, _ := io.ReadAll(response.Body)
		te = multierr.Append(te, errors.New(string(body)))
		return multierr.Append(te, fmt.Errorf("failed decode response in %s, %w", url, err))
	}

	return nil
}
