package model

import "time"

type Document struct {
	ID          string    `gorm:"primaryKey;size:36" json:"id"`
	Title       string    `gorm:"size:500;not null" json:"title"`
	FileName    string    `gorm:"size:500;not null" json:"file_name"`
	FilePath       string `gorm:"size:1000;not null" json:"-"`
	OriginalPath   string `gorm:"size:1000" json:"-"`
	SignedFilePath string `gorm:"size:1000" json:"-"`
	FileSize       int64  `gorm:"not null" json:"file_size"`
	ContactID   string    `gorm:"size:36;index" json:"contact_id"`
	ContactName string    `gorm:"size:255" json:"contact_name"`
	UploadedBy  string    `gorm:"size:36;not null;index" json:"uploaded_by"`
	Status      string    `gorm:"size:30;default:pending;index" json:"status"`
	SignedAt    *time.Time `json:"signed_at"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	Signatures []Signature    `gorm:"foreignKey:DocumentID" json:"signatures,omitempty"`
	AuditLogs  []SignatureLog `gorm:"foreignKey:DocumentID" json:"audit_logs,omitempty"`
}

type Signature struct {
	ID         string    `gorm:"primaryKey;size:36" json:"id"`
	DocumentID string    `gorm:"size:36;not null;index" json:"document_id"`
	SignerName string    `gorm:"size:255;not null" json:"signer_name"`
	SignerEmail string   `gorm:"size:255" json:"signer_email"`
	SignerID   string    `gorm:"size:36;index" json:"signer_id"`
	ImagePath  string    `gorm:"size:1000;not null" json:"-"`
	IPAddress  string    `gorm:"size:45" json:"ip_address"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type SignatureLog struct {
	ID         string    `gorm:"primaryKey;size:36" json:"id"`
	DocumentID string    `gorm:"size:36;not null;index" json:"document_id"`
	Action     string    `gorm:"size:50;not null" json:"action"`
	ActorName  string    `gorm:"size:255" json:"actor_name"`
	ActorID    string    `gorm:"size:36" json:"actor_id"`
	IPAddress  string    `gorm:"size:45" json:"ip_address"`
	Details    string    `gorm:"type:text" json:"details"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}
