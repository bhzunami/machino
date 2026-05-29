package model

import "time"

type User struct {
	ID         string    `json:"id"`
	Email      string    `json:"email"`
	Name       string    `json:"name"`
	Searchable bool      `json:"searchable"`
	CreatedAt  time.Time `json:"createdAt"`
}

type UserSearchResult struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Project struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	Favorite    bool      `json:"favorite"`
	IsOwner     bool      `json:"isOwner"`
	MemberCount int       `json:"memberCount"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type ProjectMember struct {
	UserID   string    `json:"userId"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joinedAt"`
}

type Todo struct {
	ID          string     `json:"id"`
	ProjectID   string     `json:"projectId"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	DueDate     *time.Time `json:"dueDate,omitempty"`
	Priority    string     `json:"priority"`
	Completed   bool       `json:"completed"`
	Position    int        `json:"position"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}
