package models

import "time"

type Workspace struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	LastOpenedAt time.Time `json:"lastOpenedAt"`
	CreatedAt    time.Time `json:"createdAt"`
}

type WorkspaceCreate struct {
	Name              string `json:"name" binding:"required"`
	Path              string `json:"path" binding:"required"`
	OwnerDisplayName  string `json:"ownerDisplayName,omitempty"`
}

type WorkspaceUpdate struct {
	Name string `json:"name,omitempty"`
	Path string `json:"path,omitempty"`
}