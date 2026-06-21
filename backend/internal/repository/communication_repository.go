package repository

import (
	"context"
	"errors"
	"math"
	"time"

	"gorm.io/gorm"

	"backend/internal/model"
)

type CommunicationRepository interface {
	CreateThread(ctx context.Context, t *model.EmailThread) error
	FindThreadByID(ctx context.Context, id string) (*model.EmailThread, error)
	FindThreadByContactAndSubject(ctx context.Context, contactID, subject string) (*model.EmailThread, error)
	UpdateThread(ctx context.Context, t *model.EmailThread) error
	ListThreads(ctx context.Context, search, status, direction, contactID string, page, pageSize int) ([]*model.EmailThread, int64, error)
	DeleteThread(ctx context.Context, id string) error

	CreateMessage(ctx context.Context, m *model.EmailMessage) error
	FindMessageByID(ctx context.Context, id string) (*model.EmailMessage, error)
	FindMessageByMessageID(ctx context.Context, msgID string) (*model.EmailMessage, error)
	ListMessages(ctx context.Context, threadID string) ([]*model.EmailMessage, error)
	MarkMessagesRead(ctx context.Context, threadID string) error
	StarMessage(ctx context.Context, msgID string, starred bool) error

	CreateTemplate(ctx context.Context, t *model.EmailTemplate) error
	FindTemplateByID(ctx context.Context, id string) (*model.EmailTemplate, error)
	UpdateTemplate(ctx context.Context, t *model.EmailTemplate) error
	DeleteTemplate(ctx context.Context, id string) error
	ListTemplates(ctx context.Context) ([]*model.EmailTemplate, error)
	IncrementTemplateUsage(ctx context.Context, id string) error

	CreateBulkJob(ctx context.Context, j *model.BulkEmailJob) error
	FindBulkJobByID(ctx context.Context, id string) (*model.BulkEmailJob, error)
	UpdateBulkJob(ctx context.Context, j *model.BulkEmailJob) error
	ListBulkJobs(ctx context.Context) ([]*model.BulkEmailJob, error)

	CreateSignature(ctx context.Context, s *model.EmailSignature) error
	FindSignatureByID(ctx context.Context, id string) (*model.EmailSignature, error)
	UpdateSignature(ctx context.Context, s *model.EmailSignature) error
	DeleteSignature(ctx context.Context, id string) error
	ListSignatures(ctx context.Context, userID string) ([]*model.EmailSignature, error)
	ClearDefaultSignatures(ctx context.Context, userID string) error

	Stats(ctx context.Context) (map[string]int64, error)
	RecentContacts(ctx context.Context, limit int) ([]*model.EmailThread, error)

	CreateEmailAccount(ctx context.Context, a *model.EmailAccount) error
	FindEmailAccountByUserID(ctx context.Context, userID string) (*model.EmailAccount, error)
	UpdateEmailAccount(ctx context.Context, a *model.EmailAccount) error
	DeleteEmailAccount(ctx context.Context, id string) error
	ListActiveEmailAccounts(ctx context.Context) ([]*model.EmailAccount, error)

	FindContactEmail(ctx context.Context, contactID string) (string, error)
	FindContactByID(ctx context.Context, contactID string) (*model.Contact, error)
}

type communicationRepository struct {
	db *gorm.DB
}

func NewCommunicationRepository(db *gorm.DB) CommunicationRepository {
	return &communicationRepository{db: db}
}

func (r *communicationRepository) CreateThread(ctx context.Context, t *model.EmailThread) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *communicationRepository) FindThreadByID(ctx context.Context, id string) (*model.EmailThread, error) {
	var t model.EmailThread
	err := r.db.WithContext(ctx).Preload("Messages", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at ASC")
	}).First(&t, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (r *communicationRepository) FindThreadByContactAndSubject(ctx context.Context, contactID, subject string) (*model.EmailThread, error) {
	var t model.EmailThread
	err := r.db.WithContext(ctx).Where("contact_id = ? AND subject = ?", contactID, subject).First(&t).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (r *communicationRepository) UpdateThread(ctx context.Context, t *model.EmailThread) error {
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *communicationRepository) ListThreads(ctx context.Context, search, status, direction, contactID string, page, pageSize int) ([]*model.EmailThread, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.EmailThread{})

	if search != "" {
		like := "%" + search + "%"
		q = q.Where("subject LIKE ? OR contact_name LIKE ? OR contact_email LIKE ?", like, like, like)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if direction != "" {
		q = q.Where("direction = ?", direction)
	}
	if contactID != "" {
		q = q.Where("contact_id = ?", contactID)
	}

	var total int64
	cq := r.db.WithContext(ctx).Model(&model.EmailThread{})
	if search != "" {
		like := "%" + search + "%"
		cq = cq.Where("subject LIKE ? OR contact_name LIKE ? OR contact_email LIKE ?", like, like, like)
	}
	if status != "" {
		cq = cq.Where("status = ?", status)
	}
	if direction != "" {
		cq = cq.Where("direction = ?", direction)
	}
	if contactID != "" {
		cq = cq.Where("contact_id = ?", contactID)
	}
	cq.Count(&total)

	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	var threads []*model.EmailThread
	err := q.Order("CASE WHEN unread_count > 0 THEN 0 ELSE 1 END, last_message_at DESC").Limit(pageSize).Offset(offset).Find(&threads).Error
	return threads, total, err
}

func (r *communicationRepository) DeleteThread(ctx context.Context, id string) error {
	r.db.WithContext(ctx).Where("thread_id = ?", id).Delete(&model.EmailMessage{})
	return r.db.WithContext(ctx).Delete(&model.EmailThread{}, "id = ?", id).Error
}

func (r *communicationRepository) CreateMessage(ctx context.Context, m *model.EmailMessage) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *communicationRepository) FindMessageByID(ctx context.Context, id string) (*model.EmailMessage, error) {
	var m model.EmailMessage
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *communicationRepository) FindMessageByMessageID(ctx context.Context, msgID string) (*model.EmailMessage, error) {
	var m model.EmailMessage
	err := r.db.WithContext(ctx).Where("message_id = ?", msgID).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *communicationRepository) ListMessages(ctx context.Context, threadID string) ([]*model.EmailMessage, error) {
	var msgs []*model.EmailMessage
	err := r.db.WithContext(ctx).Where("thread_id = ?", threadID).Order("created_at ASC").Find(&msgs).Error
	return msgs, err
}

func (r *communicationRepository) MarkMessagesRead(ctx context.Context, threadID string) error {
	err := r.db.WithContext(ctx).Model(&model.EmailMessage{}).
		Where("thread_id = ? AND is_read = false AND direction = 'inbound'", threadID).
		Update("is_read", true).Error
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&model.EmailThread{}).
		Where("id = ?", threadID).
		Update("unread_count", 0).Error
}

func (r *communicationRepository) StarMessage(ctx context.Context, msgID string, starred bool) error {
	return r.db.WithContext(ctx).Model(&model.EmailMessage{}).Where("id = ?", msgID).Update("is_starred", starred).Error
}

func (r *communicationRepository) CreateTemplate(ctx context.Context, t *model.EmailTemplate) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *communicationRepository) FindTemplateByID(ctx context.Context, id string) (*model.EmailTemplate, error) {
	var t model.EmailTemplate
	err := r.db.WithContext(ctx).First(&t, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (r *communicationRepository) UpdateTemplate(ctx context.Context, t *model.EmailTemplate) error {
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *communicationRepository) DeleteTemplate(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.EmailTemplate{}, "id = ?", id).Error
}

func (r *communicationRepository) ListTemplates(ctx context.Context) ([]*model.EmailTemplate, error) {
	var templates []*model.EmailTemplate
	err := r.db.WithContext(ctx).Order("used_count DESC, name ASC").Find(&templates).Error
	return templates, err
}

func (r *communicationRepository) IncrementTemplateUsage(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&model.EmailTemplate{}).Where("id = ?", id).
		UpdateColumn("used_count", gorm.Expr("used_count + 1")).Error
}

func (r *communicationRepository) CreateBulkJob(ctx context.Context, j *model.BulkEmailJob) error {
	return r.db.WithContext(ctx).Create(j).Error
}

func (r *communicationRepository) FindBulkJobByID(ctx context.Context, id string) (*model.BulkEmailJob, error) {
	var j model.BulkEmailJob
	err := r.db.WithContext(ctx).First(&j, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &j, nil
}

func (r *communicationRepository) UpdateBulkJob(ctx context.Context, j *model.BulkEmailJob) error {
	return r.db.WithContext(ctx).Save(j).Error
}

func (r *communicationRepository) ListBulkJobs(ctx context.Context) ([]*model.BulkEmailJob, error) {
	var jobs []*model.BulkEmailJob
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&jobs).Error
	return jobs, err
}

func (r *communicationRepository) CreateSignature(ctx context.Context, s *model.EmailSignature) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *communicationRepository) FindSignatureByID(ctx context.Context, id string) (*model.EmailSignature, error) {
	var s model.EmailSignature
	err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *communicationRepository) UpdateSignature(ctx context.Context, s *model.EmailSignature) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *communicationRepository) DeleteSignature(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.EmailSignature{}, "id = ?", id).Error
}

func (r *communicationRepository) ListSignatures(ctx context.Context, userID string) ([]*model.EmailSignature, error) {
	var sigs []*model.EmailSignature
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("is_default DESC, name ASC").Find(&sigs).Error
	return sigs, err
}

func (r *communicationRepository) ClearDefaultSignatures(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Model(&model.EmailSignature{}).Where("user_id = ?", userID).Update("is_default", false).Error
}

func (r *communicationRepository) Stats(ctx context.Context) (map[string]int64, error) {
	stats := make(map[string]int64)
	var total, open, unread, bulkJobs int64
	r.db.WithContext(ctx).Model(&model.EmailThread{}).Count(&total)
	r.db.WithContext(ctx).Model(&model.EmailThread{}).Where("status = 'open'").Count(&open)
	r.db.WithContext(ctx).Model(&model.EmailMessage{}).Where("is_read = false AND direction = 'inbound'").Count(&unread)
	r.db.WithContext(ctx).Model(&model.BulkEmailJob{}).Where("status = 'completed'").Count(&bulkJobs)

	today := time.Now().Truncate(24 * time.Hour)
	var sentToday int64
	r.db.WithContext(ctx).Model(&model.EmailMessage{}).Where("direction = 'outbound' AND created_at >= ?", today).Count(&sentToday)

	stats["total_threads"] = total
	stats["open_threads"] = open
	stats["unread_messages"] = unread
	stats["sent_today"] = sentToday
	stats["bulk_jobs_run"] = bulkJobs
	_ = math.Floor(0)
	return stats, nil
}

func (r *communicationRepository) RecentContacts(ctx context.Context, limit int) ([]*model.EmailThread, error) {
	var threads []*model.EmailThread
	err := r.db.WithContext(ctx).
		Where("contact_id != ''").
		Order("last_message_at DESC").
		Limit(limit).
		Find(&threads).Error
	return threads, err
}

func (r *communicationRepository) CreateEmailAccount(ctx context.Context, a *model.EmailAccount) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *communicationRepository) FindEmailAccountByUserID(ctx context.Context, userID string) (*model.EmailAccount, error) {
	var a model.EmailAccount
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&a).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (r *communicationRepository) UpdateEmailAccount(ctx context.Context, a *model.EmailAccount) error {
	return r.db.WithContext(ctx).Save(a).Error
}

func (r *communicationRepository) DeleteEmailAccount(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.EmailAccount{}, "id = ?", id).Error
}

func (r *communicationRepository) ListActiveEmailAccounts(ctx context.Context) ([]*model.EmailAccount, error) {
	var accounts []*model.EmailAccount
	err := r.db.WithContext(ctx).Where("is_active = ? AND imap_host != ''", true).Find(&accounts).Error
	return accounts, err
}

func (r *communicationRepository) FindContactEmail(ctx context.Context, contactID string) (string, error) {
	var email string
	err := r.db.WithContext(ctx).Model(&model.Contact{}).Select("email").Where("id = ?", contactID).Scan(&email).Error
	return email, err
}

func (r *communicationRepository) FindContactByID(ctx context.Context, contactID string) (*model.Contact, error) {
	var c model.Contact
	err := r.db.WithContext(ctx).Where("id = ?", contactID).First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}
