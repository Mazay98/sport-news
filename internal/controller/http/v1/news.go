package v1

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
	"go.sport-news/internal/repository"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type response struct {
	Status   status      `json:"status"`
	Data     interface{} `json:"data,omitempty"`
	Message  string      `json:"message,omitempty"`
	Metadata meta        `json:"metadata,omitempty"`
}
type responseCache struct {
	status   int
	response response
}

type status string
type meta struct {
	CreatedAt  string  `json:"createdAt"`
	TotalItems *int    `json:"totalItems,omitempty"`
	Sort       *string `json:"sort,omitempty"`
}

const (
	success status = "success"
	errors  status = "error"
)
const timeFormat string = "2006-01-02T15:04:05Z"

type NewsController struct {
	newsRepository repository.NewsRepository
	cache          *cache.Cache
	logger         *zap.Logger
}

type INewsController interface {
	GetTeamNews(w http.ResponseWriter, r *http.Request)
	GetTeamNewsByID(w http.ResponseWriter, r *http.Request)
	ResetCache(w http.ResponseWriter, r *http.Request)
}

func NewNewsController(newsRepository repository.NewsRepository, cache *cache.Cache, logger *zap.Logger) *NewsController {
	return &NewsController{newsRepository: newsRepository, cache: cache, logger: logger}
}

// ResetCache handle GET /v1/cache-flush - reset cache.
func (c *NewsController) ResetCache(w http.ResponseWriter, r *http.Request) {
	c.cache.Flush()

	c.respondWithJSON(w, responseCache{
		status: http.StatusOK,
		response: response{
			Status:   success,
			Data:     nil,
			Message:  "success",
			Metadata: meta{},
		},
	})
}

// GetTeamNews handle GET /v1/teams/{team}/news.
func (c *NewsController) GetTeamNews(w http.ResponseWriter, r *http.Request) {
	resp, found := c.cache.Get(r.URL.String())
	if !found {
		vars := mux.Vars(r)
		team := vars["team"]

		if ok := c.validateExist(w, r, "teamId", team); !ok {
			return
		}

		a, err := c.newsRepository.GetTeamNews(r.Context(), team)
		if err != nil {
			c.logger.Error("failed get getTeamNews", zap.String("url", r.URL.String()))
			c.internalErrorResponse(w)
			return
		}
		ti := len(*a)
		s := "-published"

		resp = responseCache{
			status: http.StatusOK,
			response: response{
				Status: success,
				Data:   &a,
				Metadata: meta{
					CreatedAt:  time.Now().Format(timeFormat),
					TotalItems: &ti,
					Sort:       &s,
				},
			},
		}
		c.cache.Set(r.URL.String(), resp, cache.DefaultExpiration)
	}

	c.respondWithJSON(w, resp.(responseCache))
}

// GetTeamNewsByID handle GET /v1/teams/{team}/news/{id}.
func (c *NewsController) GetTeamNewsByID(w http.ResponseWriter, r *http.Request) {
	resp, found := c.cache.Get(r.URL.String())
	if !found {
		vars := mux.Vars(r)
		team := vars["team"]
		id := vars["id"]

		if ok := c.validateExist(w, r, "teamId", team); !ok {
			return
		}
		if ok := c.validateExist(w, r, "id", id); !ok {
			return
		}

		a, err := c.newsRepository.GetTeamNewsByID(r.Context(), team, id)
		if err != nil {
			c.logger.Error("not found GetTeamNewsByID",
				zap.String("url", r.URL.String()),
				zap.String("team", team),
				zap.String("uuid", id),
			)
			resp = responseCache{
				status: http.StatusNotFound,
				response: response{
					Status: errors,
					Data:   &a,
					Metadata: meta{
						CreatedAt: time.Now().Format(timeFormat),
					},
				},
			}
			c.cache.Set(r.URL.String(), resp, cache.DefaultExpiration)
			c.respondWithJSON(w, resp.(responseCache))
			return
		}

		resp = responseCache{
			status: http.StatusOK,
			response: response{
				Status: success,
				Data:   &a,
				Metadata: meta{
					CreatedAt: time.Now().Format(timeFormat),
				},
			},
		}
		c.cache.Set(r.URL.String(), resp, cache.DefaultExpiration)

	}

	c.respondWithJSON(w, resp.(responseCache))
}

func (c *NewsController) respondWithJSON(w http.ResponseWriter, resp responseCache) {
	response, err := json.Marshal(resp.response)
	if err != nil {
		c.internalErrorResponse(w)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.status)
	_, err = w.Write(response)
	if err != nil {
		c.logger.Error("failed send json response", zap.Error(err))
	}
}

func (c *NewsController) internalErrorResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	_, err := w.Write([]byte(fmt.Sprintf("{\"status\":\"%s\",\"message\":\"%s\"}", errors, "internal server error")))
	if err != nil {
		c.logger.Error("failed send internal error response", zap.Error(err))
	}
}

// validateExist check exist field with value ond db.
func (c *NewsController) validateExist(w http.ResponseWriter, r *http.Request, field string, value any) bool {
	exist, err := c.newsRepository.Exist(r.Context(), field, value)
	if err != nil || !exist {
		resp := responseCache{
			status: http.StatusNotFound,
			response: response{
				Status:  errors,
				Message: fmt.Sprintf("%s not found", field),
				Metadata: meta{
					CreatedAt: time.Now().Format(timeFormat),
				},
			},
		}

		c.cache.Set(r.URL.String(), resp, cache.DefaultExpiration)
		c.respondWithJSON(w, resp)
		return false
	}
	return true
}
