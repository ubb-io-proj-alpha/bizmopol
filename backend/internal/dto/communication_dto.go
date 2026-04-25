package dto

import "time"

type ThreadListResponse struct {
	Data       []ThreadResponse `json:"data"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

type ThreadResponse struct {
	ID           string          `json:"id"`
	Subject      string          `json:"subject"`
	ContactID    string          `json:"contact_id"`
	ContactName  string          `json:"contact_name"`
	ContactEmail string          `json:"contact_email"`
	Status       string          `json:"status"`
	Direction    string          `json:"direction"`
	LastMessageAt time.Time      `json:"last_message_at"`
	MessageCount  int            `json:"message_count"`
	UnreadCount   int            `json:"unread_count"`
	Labels       string          `json:"labels"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	Messages     []MessageResponse `json:"messages,omitempty"`
}

type MessageResponse struct {
	ID        string     `json:"id"`
	ThreadID  string     `json:"thread_id"`
	MessageID string     `json:"message_id"`
	From      string     `json:"from"`
	To        string     `json:"to"`
	Cc        string     `json:"cc"`
	Subject   string     `json:"subject"`
	Body      string     `json:"body"`
	BodyHTML  string     `json:"body_html"`
	Direction string     `json:"direction"`
	IsRead    bool       `json:"is_read"`
	IsStarred bool       `json:"is_starred"`
	SentAt    *time.Time `json:"sent_at"`
	CreatedAt time.Time  `json:"created_at"`
}

type SendEmailRequest struct {
	ContactID  string   `json:"contact_id" binding:"required"`
	To         string   `json:"to" binding:"required"`
	Subject    string   `json:"subject" binding:"required"`
	Body       string   `json:"body" binding:"required"`
	BodyHTML   string   `json:"body_html"`
	Cc         []string `json:"cc"`
	ThreadID   string   `json:"thread_id"`
	TemplateID string   `json:"template_id"`
	SignatureID string  `json:"signature_id"`
}

type ReplyEmailRequest struct {
	Body        string   `json:"body" binding:"required"`
	BodyHTML    string   `json:"body_html"`
	Cc          []string `json:"cc"`
	SignatureID string   `json:"signature_id"`
}

type BulkEmailRequest struct {
	Name        string   `json:"name" binding:"required"`
	Subject     string   `json:"subject" binding:"required"`
	Body        string   `json:"body" binding:"required"`
	BodyHTML    string   `json:"body_html"`
	ContactIDs  []string `json:"contact_ids" binding:"required,min=1"`
	TemplateID  string   `json:"template_id"`
	ScheduledAt *time.Time `json:"scheduled_at"`
}

type BulkEmailResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Subject     string     `json:"subject"`
	Status      string     `json:"status"`
	TotalCount  int        `json:"total_count"`
	SentCount   int        `json:"sent_count"`
	FailedCount int        `json:"failed_count"`
	ScheduledAt *time.Time `json:"scheduled_at"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

type TemplateCreateRequest struct {
	Name     string `json:"name" binding:"required"`
	Subject  string `json:"subject" binding:"required"`
	Body     string `json:"body" binding:"required"`
	BodyHTML string `json:"body_html"`
	Category string `json:"category"`
}

type TemplateUpdateRequest struct {
	Name     string `json:"name"`
	Subject  string `json:"subject"`
	Body     string `json:"body"`
	BodyHTML string `json:"body_html"`
	Category string `json:"category"`
}

type TemplateResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	BodyHTML  string    `json:"body_html"`
	Category  string    `json:"category"`
	UsedCount int       `json:"used_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SignatureCreateRequest struct {
	Name      string `json:"name" binding:"required"`
	Body      string `json:"body" binding:"required"`
	IsDefault bool   `json:"is_default"`
}

type SignatureUpdateRequest struct {
	Name      string `json:"name"`
	Body      string `json:"body"`
	IsDefault bool   `json:"is_default"`
}

type SignatureResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Body      string    `json:"body"`
	IsDefault bool      `json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
}

type ThreadQuery struct {
	Search    string `form:"search"`
	Status    string `form:"status"`
	Direction string `form:"direction"`
	ContactID string `form:"contact_id"`
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
}

type SyncStatusResponse struct {
	Connected    bool      `json:"connected"`
	LastSync     *time.Time `json:"last_sync"`
	EmailAddress string    `json:"email_address"`
	Error        string    `json:"error,omitempty"`
}

type StatsResponse struct {
	TotalThreads   int64 `json:"total_threads"`
	OpenThreads    int64 `json:"open_threads"`
	UnreadMessages int64 `json:"unread_messages"`
	SentToday      int64 `json:"sent_today"`
	BulkJobsRun    int64 `json:"bulk_jobs_run"`
	RecentContacts []RecentContactResponse `json:"recent_contacts"`
}

type RecentContactResponse struct {
	ContactID    string    `json:"contact_id"`
	ContactName  string    `json:"contact_name"`
	ContactEmail string    `json:"contact_email"`
	LastContact  time.Time `json:"last_contact"`
	ThreadCount  int       `json:"thread_count"`
}
