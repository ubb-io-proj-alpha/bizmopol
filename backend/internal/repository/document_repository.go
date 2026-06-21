package repository

import (
	"context"

	"backend/internal/model"

	"gorm.io/gorm"
)

type DocumentRepository interface {
	CreateDocument(ctx context.Context, doc *model.Document) error
	GetDocument(ctx context.Context, id string) (*model.Document, error)
	ListDocuments(ctx context.Context, search, status, contactID string, page, pageSize int) ([]*model.Document, int64, error)
	DeleteDocument(ctx context.Context, id string) error
	UpdateDocument(ctx context.Context, doc *model.Document) error

	CreateSignature(ctx context.Context, sig *model.Signature) error
	GetSignature(ctx context.Context, id string) (*model.Signature, error)
	ListSignaturesByDocument(ctx context.Context, documentID string) ([]*model.Signature, error)

	CreateLog(ctx context.Context, log *model.SignatureLog) error
	ListLogsByDocument(ctx context.Context, documentID string) ([]*model.SignatureLog, error)
}

type documentRepository struct {
	db *gorm.DB
}

func NewDocumentRepository(db *gorm.DB) DocumentRepository {
	return &documentRepository{db: db}
}

func (r *documentRepository) CreateDocument(ctx context.Context, doc *model.Document) error {
	return r.db.WithContext(ctx).Create(doc).Error
}

func (r *documentRepository) GetDocument(ctx context.Context, id string) (*model.Document, error) {
	var doc model.Document
	err := r.db.WithContext(ctx).Preload("Signatures").Preload("AuditLogs", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC")
	}).First(&doc, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &doc, err
}

func (r *documentRepository) ListDocuments(ctx context.Context, search, status, contactID string, page, pageSize int) ([]*model.Document, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Document{})

	if search != "" {
		like := "%" + search + "%"
		q = q.Where("title LIKE ? OR file_name LIKE ? OR contact_name LIKE ?", like, like, like)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if contactID != "" {
		q = q.Where("contact_id = ?", contactID)
	}

	var total int64
	q.Count(&total)

	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	var docs []*model.Document
	err := q.Preload("Signatures").Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&docs).Error
	return docs, total, err
}

func (r *documentRepository) DeleteDocument(ctx context.Context, id string) error {
	r.db.WithContext(ctx).Where("document_id = ?", id).Delete(&model.SignatureLog{})
	r.db.WithContext(ctx).Where("document_id = ?", id).Delete(&model.Signature{})
	return r.db.WithContext(ctx).Delete(&model.Document{}, "id = ?", id).Error
}

func (r *documentRepository) UpdateDocument(ctx context.Context, doc *model.Document) error {
	return r.db.WithContext(ctx).Save(doc).Error
}

func (r *documentRepository) CreateSignature(ctx context.Context, sig *model.Signature) error {
	return r.db.WithContext(ctx).Create(sig).Error
}

func (r *documentRepository) GetSignature(ctx context.Context, id string) (*model.Signature, error) {
	var sig model.Signature
	err := r.db.WithContext(ctx).First(&sig, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &sig, err
}

func (r *documentRepository) ListSignaturesByDocument(ctx context.Context, documentID string) ([]*model.Signature, error) {
	var sigs []*model.Signature
	err := r.db.WithContext(ctx).Where("document_id = ?", documentID).Order("created_at ASC").Find(&sigs).Error
	return sigs, err
}

func (r *documentRepository) CreateLog(ctx context.Context, log *model.SignatureLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *documentRepository) ListLogsByDocument(ctx context.Context, documentID string) ([]*model.SignatureLog, error) {
	var logs []*model.SignatureLog
	err := r.db.WithContext(ctx).Where("document_id = ?", documentID).Order("created_at DESC").Find(&logs).Error
	return logs, err
}
