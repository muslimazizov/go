package good

import "time"

// Good model
type Good struct {
	Id          int64     `json:"id"`
	ProjectId   int64     `json:"projectId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Priority    int64     `json:"priority"`
	Removed     bool      `json:"removed"`
	CreatedAt   time.Time `json:"createdAt"`
}

// Project model
type Project struct {
	Id        int64
	Name      string
	CreatedAt time.Time
}
