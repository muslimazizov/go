package good

import "time"

type QueryParams struct {
	Id        int64 `schema:"id,required"`
	ProjectId int64 `schema:"projectId,required"`
}

type ListGoodsParams struct {
	Limit  int64
	Offset int64
}

type ListGoodsMeta struct {
	Total   int64 `json:"total"`
	Removed int64 `json:"removed"`
	Limit   int64 `json:"limit"`
	Offset  int64 `json:"offset"`
}

type ListGoodsResponse struct {
	Meta  ListGoodsMeta `json:"meta"`
	Goods []Good        `json:"goods"`
}

type CreateGoodParams struct {
	Id          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ProjectId   int64
}

type UpdateGoodParams struct {
	ProjectId   int64     `json:"projectId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Priority    int64     `json:"priority"`
	Removed     *bool     `json:"removed"`
	CreatedAt   time.Time `json:"createdAt"`
}

type ReprioritizeGoodParams struct {
	NewPriority int64 `json:"newPriority"`
}

type ReprioritizedGood struct {
	Id       int64 `json:"id"`
	Priority int64 `json:"priority"`
}
