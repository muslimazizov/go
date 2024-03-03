package good

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/schema"
	"github.com/jackc/pgx/v4"
)

func init() {
	decoder.IgnoreUnknownKeys(true)
}

var decoder = schema.NewDecoder()

type ErrorResponse struct {
	Code    int64             `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details"`
}

func ListGoods(s Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var params struct {
			Limit  int64 `schema:"limit"`
			Offset int64 `schema:"offset"`
		}

		if err := decoder.Decode(&params, r.URL.Query()); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var limit int64 = 10
		var offset int64 = 1
		if params.Limit > 0 {
			limit = params.Limit
		}

		if params.Offset > 0 {
			offset = params.Offset
		}

		goods, err := s.ListGoods(r.Context(), ListGoodsParams{
			Limit:  limit,
			Offset: offset,
		})

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("failed to list goods: %v", err)
			return
		}

		w.WriteHeader(http.StatusOK)
		if goods == nil {
			goods = make([]Good, 0)
		}

		response := ListGoodsResponse{
			Meta: ListGoodsMeta{
				Total:   s.store.Count(r.Context()),
				Removed: s.store.RemovedCount(r.Context()),
				Limit:   limit,
				Offset:  offset,
			},
			Goods: goods,
		}

		_ = json.NewEncoder(w).Encode(response)
	}
}

func CreateGood(s Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		projectId, err := strconv.Atoi(r.URL.Query().Get("projectId"))
		if r.URL.Query().Get("projectId") == "" || err != nil {
			res := make(map[string]string)
			res["error"] = "projectId is required"
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(res)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		var params CreateGoodParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if params.Name == "" {
			res := make(map[string]string)
			res["error"] = "name is required"
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(res)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		if params.Id == 0 {
			res := make(map[string]string)
			res["error"] = "id is required"
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(res)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		params.ProjectId = int64(projectId)

		result, err := s.CreateGood(context.Background(), params)
		if err != nil {
			log.Printf("failed to create good: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(result)
	}
}

func DeleteGood(s Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var params QueryParams
		if err := decoder.Decode(&params, r.URL.Query()); err != nil {
			res := make(map[string]string)
			res["error"] = "id and projectId query params are required"
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(res)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		good, err := s.DeleteGood(r.Context(), params)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				_ = json.NewEncoder(w).Encode(ErrorResponse{
					Code:    NotFoundErrorCode,
					Message: NotFoundErrorMessage,
					Details: make(map[string]string),
				})
				return
			}

			log.Printf("failed to delete good: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_ = json.NewEncoder(w).Encode(struct {
			Id        int64 `json:"id"`
			ProjectId int64 `json:"projectId"`
			Removed   bool  `json:"removed"`
		}{
			Id:        good.Id,
			ProjectId: good.ProjectId,
			Removed:   good.Removed,
		})
	}
}

func UpdateGood(s Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var queryParams QueryParams
		if err := decoder.Decode(&queryParams, r.URL.Query()); err != nil {
			res := make(map[string]string)
			res["error"] = "id and projectId query params are required"
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(res)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		var params UpdateGoodParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if params.Name == "" {
			res := make(map[string]string)
			res["error"] = "name is required"
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(res)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		if params.Priority == 0 {
			res := make(map[string]string)
			res["error"] = "priority field is required"
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(res)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		if params.Removed == nil {
			res := make(map[string]string)
			res["error"] = "removed field is required"
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(res)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		if params.ProjectId == 0 {
			res := make(map[string]string)
			res["error"] = "projectId field is required or cannot be equal to 0"
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(res)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		good, err := s.UpdateGood(r.Context(), queryParams.Id, queryParams.ProjectId, params)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				_ = json.NewEncoder(w).Encode(ErrorResponse{
					Code:    NotFoundErrorCode,
					Message: NotFoundErrorMessage,
					Details: make(map[string]string),
				})
				return
			}

			log.Printf("failed to delete good: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_ = json.NewEncoder(w).Encode(good)
	}
}

func ReprioritizeGood(s Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var queryParams QueryParams
		if err := decoder.Decode(&queryParams, r.URL.Query()); err != nil {
			res := make(map[string]string)
			res["error"] = "id and projectId query params are required"
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(res)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		var params ReprioritizeGoodParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if params.NewPriority == 0 {
			res := make(map[string]string)
			res["error"] = "newPriority field is required"
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(res)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		goods, err := s.ReprioritizeGood(r.Context(), queryParams.Id, queryParams.ProjectId, params)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				_ = json.NewEncoder(w).Encode(ErrorResponse{
					Code:    NotFoundErrorCode,
					Message: NotFoundErrorMessage,
					Details: make(map[string]string),
				})
				return
			}

			log.Printf("failed to reprioritize good: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_ = json.NewEncoder(w).Encode(struct {
			Priorities []ReprioritizedGood `json:"priorities"`
		}{
			Priorities: goods,
		})
	}
}
