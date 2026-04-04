package models

import "time"

type MemberRole string

const (
	RoleOwner      MemberRole = "owner"
	RoleAdmin      MemberRole = "admin"
	RoleSecretary  MemberRole = "secretary" // 秘书（监工/协调式监督角色）
	RoleAssistant  MemberRole = "assistant"
	RoleMember     MemberRole = "member"
)

type Member struct {
	ID                string     `json:"id"`
	WorkspaceID       string     `json:"workspaceId"`
	Name              string     `json:"name"`
	RoleType          MemberRole `json:"roleType"`
	RoleKey           string     `json:"roleKey,omitempty"`
	Avatar            string     `json:"avatar,omitempty"`
	TerminalType      string     `json:"terminalType,omitempty"`
	TerminalCommand   string     `json:"terminalCommand,omitempty"`
	TerminalPath      string     `json:"terminalPath,omitempty"`
	AutoStartTerminal bool       `json:"autoStartTerminal"`
	Status            string     `json:"status"`
	CreatedAt         time.Time  `json:"createdAt"`
}

type MemberCreate struct {
	Name            string     `json:"name" binding:"required"`
	RoleType        MemberRole `json:"roleType" binding:"required"`
	TerminalType    string     `json:"terminalType,omitempty"`
	TerminalCommand string     `json:"terminalCommand,omitempty"`
}