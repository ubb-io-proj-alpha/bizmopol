package model

import "time"

type EmailThread struct {
	ID          string         `gorm:"primaryKey;size:36" json:"id"`
	Subject     string         `gorm:"size:500;not null" json:"subject"`
	ContactID   string         `gorm:"size:36;index" json:"contact_id"`
	ContactName string         `gorm:"size:255" json:"contact_name"`
	ContactEmail string        `gorm:"size:255;index" json:"contact_email"`
	Status      string         `gorm:"size:20;default:open" json:"status"`
	Direction   string         `gorm:"size:10;default:inbound" json:"direction"`
	LastMessageAt time.Time    `gorm:"index" json:"last_message_at"`
	MessageCount  int          `gorm:"default:0" json:"message_count"`
	UnreadCount   int          `gorm:"default:0" json:"unread_count"`
	Labels      string         `gorm:"type:text" json:"labels"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	Messages    []EmailMessage `gorm:"foreignKey:ThreadID" json:"messages,omitempty"`
}

type EmailMessage struct {
	ID          string    `gorm:"primaryKey;size:36" json:"id"`
	ThreadID    string    `gorm:"size:36;not null;index" json:"thread_id"`
	MessageID   string    `gorm:"size:500;uniqueIndex" json:"message_id"`
	InReplyTo   string    `gorm:"size:500" json:"in_reply_to"`
	From        string    `gorm:"size:500;not null" json:"from"`
	To          string    `gorm:"type:text;not null" json:"to"`
	Cc          string    `gorm:"type:text" json:"cc"`
	Subject     string    `gorm:"size:500" json:"subject"`
	Body        string    `gorm:"type:text" json:"body"`
	BodyHTML    string    `gorm:"type:text" json:"body_html"`
	Direction   string    `gorm:"size:10;not null" json:"direction"`
	IsRead      bool      `gorm:"default:false" json:"is_read"`
	IsStarred   bool      `gorm:"default:false" json:"is_starred"`
	SentAt      *time.Time `gorm:"index" json:"sent_at"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type EmailTemplate struct {
	ID        string    `gorm:"primaryKey;size:36" json:"id"`
	Name      string    `gorm:"size:255;not null" json:"name"`
	Subject   string    `gorm:"size:500;not null" json:"subject"`
	Body      string    `gorm:"type:text;not null" json:"body"`
	BodyHTML  string    `gorm:"type:text" json:"body_html"`
	Category  string    `gorm:"size:100" json:"category"`
	UsedCount int       `gorm:"default:0" json:"used_count"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type BulkEmailJob struct {
	ID           string    `gorm:"primaryKey;size:36" json:"id"`
	Name         string    `gorm:"size:255;not null" json:"name"`
	Subject      string    `gorm:"size:500;not null" json:"subject"`
	Body         string    `gorm:"type:text;not null" json:"body"`
	BodyHTML     string    `gorm:"type:text" json:"body_html"`
	TemplateID   string    `gorm:"size:36" json:"template_id"`
	ContactIDs   string    `gorm:"type:text" json:"contact_ids"`
	Status       string    `gorm:"size:20;default:pending" json:"status"`
	TotalCount   int       `gorm:"default:0" json:"total_count"`
	SentCount    int       `gorm:"default:0" json:"sent_count"`
	FailedCount  int       `gorm:"default:0" json:"failed_count"`
	ScheduledAt  *time.Time `json:"scheduled_at"`
	StartedAt    *time.Time `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at"`
	ErrorLog     string    `gorm:"type:text" json:"error_log"`
	UserID       string    `gorm:"size:36" json:"user_id"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type EmailSignature struct {
	ID        string    `gorm:"primaryKey;size:36" json:"id"`
	UserID    string    `gorm:"size:36;not null;index" json:"user_id"`
	Name      string    `gorm:"size:255;not null" json:"name"`
	Body      string    `gorm:"type:text;not null" json:"body"`
	IsDefault bool      `gorm:"default:false" json:"is_default"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type EmailAccount struct {
	ID          string     `gorm:"primaryKey;size:36" json:"id"`
	UserID      string     `gorm:"size:36;not null;uniqueIndex" json:"user_id"`
	Provider    string     `gorm:"size:50;default:gmail" json:"provider"`
	Email       string     `gorm:"size:255;not null" json:"email"`
	Password    string     `gorm:"size:500" json:"-"`
	IMAPHost    string     `gorm:"size:255" json:"imap_host"`
	IMAPPort    int        `gorm:"default:993" json:"imap_port"`
	SMTPHost    string     `gorm:"size:255" json:"smtp_host"`
	SMTPPort    int        `gorm:"default:587" json:"smtp_port"`
	LastSyncUID uint32     `gorm:"default:0" json:"last_sync_uid"`
	LastSyncAt  *time.Time `json:"last_sync_at"`
	SyncedCount int        `gorm:"default:0" json:"synced_count"`
	IsActive    bool       `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}
