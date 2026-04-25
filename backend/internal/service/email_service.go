package service

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"

	"backend/internal/dto"
	"backend/internal/model"
	"backend/internal/repository"
)

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPass     string
	IMAPHost     string
	IMAPPort     string
	FromName     string
	FromEmail    string
}

func LoadEmailConfig() EmailConfig {
	return EmailConfig{
		SMTPHost:  getEmailEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:  getEmailEnv("SMTP_PORT", "587"),
		SMTPUser:  getEmailEnv("SMTP_USER", ""),
		SMTPPass:  getEmailEnv("SMTP_PASS", ""),
		IMAPHost:  getEmailEnv("IMAP_HOST", "imap.gmail.com"),
		IMAPPort:  getEmailEnv("IMAP_PORT", "993"),
		FromName:  getEmailEnv("EMAIL_FROM_NAME", "BizmoPol CRM"),
		FromEmail: getEmailEnv("EMAIL_FROM", ""),
	}
}

func getEmailEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

type CommunicationService interface {
	SendEmail(ctx context.Context, userID string, req dto.SendEmailRequest) (*dto.ThreadResponse, error)
	ReplyToThread(ctx context.Context, userID, threadID string, req dto.ReplyEmailRequest) (*dto.MessageResponse, error)
	ListThreads(ctx context.Context, q dto.ThreadQuery) (*dto.ThreadListResponse, error)
	GetThread(ctx context.Context, id string) (*dto.ThreadResponse, error)
	DeleteThread(ctx context.Context, id string) error
	MarkThreadRead(ctx context.Context, id string) error
	ArchiveThread(ctx context.Context, id string) error
	StarMessage(ctx context.Context, msgID string, starred bool) error
	UpdateThreadStatus(ctx context.Context, id, status string) error

	SendBulkEmail(ctx context.Context, userID string, req dto.BulkEmailRequest) (*dto.BulkEmailResponse, error)
	ListBulkJobs(ctx context.Context) ([]*dto.BulkEmailResponse, error)
	GetBulkJob(ctx context.Context, id string) (*dto.BulkEmailResponse, error)

	ListTemplates(ctx context.Context) ([]*dto.TemplateResponse, error)
	CreateTemplate(ctx context.Context, req dto.TemplateCreateRequest) (*dto.TemplateResponse, error)
	UpdateTemplate(ctx context.Context, id string, req dto.TemplateUpdateRequest) (*dto.TemplateResponse, error)
	DeleteTemplate(ctx context.Context, id string) error

	ListSignatures(ctx context.Context, userID string) ([]*dto.SignatureResponse, error)
	CreateSignature(ctx context.Context, userID string, req dto.SignatureCreateRequest) (*dto.SignatureResponse, error)
	UpdateSignature(ctx context.Context, id, userID string, req dto.SignatureUpdateRequest) (*dto.SignatureResponse, error)
	DeleteSignature(ctx context.Context, id string) error

	GetStats(ctx context.Context) (*dto.StatsResponse, error)
	SyncInbox(ctx context.Context) (*dto.SyncStatusResponse, error)

	GetContactThreads(ctx context.Context, contactID string) ([]*dto.ThreadResponse, error)
}

type communicationService struct {
	repo      repository.CommunicationRepository
	emailCfg  EmailConfig
}

func NewCommunicationService(repo repository.CommunicationRepository) CommunicationService {
	return &communicationService{
		repo:     repo,
		emailCfg: LoadEmailConfig(),
	}
}

func toThreadResponse(t *model.EmailThread) dto.ThreadResponse {
	msgs := make([]dto.MessageResponse, 0, len(t.Messages))
	for _, m := range t.Messages {
		msgs = append(msgs, toMessageResponse(&m))
	}
	return dto.ThreadResponse{
		ID:            t.ID,
		Subject:       t.Subject,
		ContactID:     t.ContactID,
		ContactName:   t.ContactName,
		ContactEmail:  t.ContactEmail,
		Status:        t.Status,
		Direction:     t.Direction,
		LastMessageAt: t.LastMessageAt,
		MessageCount:  t.MessageCount,
		UnreadCount:   t.UnreadCount,
		Labels:        t.Labels,
		CreatedAt:     t.CreatedAt,
		UpdatedAt:     t.UpdatedAt,
		Messages:      msgs,
	}
}

func toMessageResponse(m *model.EmailMessage) dto.MessageResponse {
	return dto.MessageResponse{
		ID:        m.ID,
		ThreadID:  m.ThreadID,
		MessageID: m.MessageID,
		From:      m.From,
		To:        m.To,
		Cc:        m.Cc,
		Subject:   m.Subject,
		Body:      m.Body,
		BodyHTML:  m.BodyHTML,
		Direction: m.Direction,
		IsRead:    m.IsRead,
		IsStarred: m.IsStarred,
		SentAt:    m.SentAt,
		CreatedAt: m.CreatedAt,
	}
}

func toBulkJobResponse(j *model.BulkEmailJob) *dto.BulkEmailResponse {
	return &dto.BulkEmailResponse{
		ID:          j.ID,
		Name:        j.Name,
		Subject:     j.Subject,
		Status:      j.Status,
		TotalCount:  j.TotalCount,
		SentCount:   j.SentCount,
		FailedCount: j.FailedCount,
		ScheduledAt: j.ScheduledAt,
		StartedAt:   j.StartedAt,
		CompletedAt: j.CompletedAt,
		CreatedAt:   j.CreatedAt,
	}
}

func toTemplateResponse(t *model.EmailTemplate) *dto.TemplateResponse {
	return &dto.TemplateResponse{
		ID:        t.ID,
		Name:      t.Name,
		Subject:   t.Subject,
		Body:      t.Body,
		BodyHTML:  t.BodyHTML,
		Category:  t.Category,
		UsedCount: t.UsedCount,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

func toSignatureResponse(s *model.EmailSignature) *dto.SignatureResponse {
	return &dto.SignatureResponse{
		ID:        s.ID,
		Name:      s.Name,
		Body:      s.Body,
		IsDefault: s.IsDefault,
		CreatedAt: s.CreatedAt,
	}
}

func (s *communicationService) sendSMTP(to []string, subject, body, bodyHTML, inReplyTo, msgID string) error {
	if s.emailCfg.SMTPUser == "" {
		slog.Warn("SMTP not configured, skipping actual email send")
		return nil
	}

	fromHeader := fmt.Sprintf("%s <%s>", s.emailCfg.FromName, s.emailCfg.FromEmail)
	toHeader := strings.Join(to, ", ")

	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("From: %s\r\n", fromHeader))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", toHeader))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString(fmt.Sprintf("Message-ID: <%s@bizmopol>\r\n", msgID))
	if inReplyTo != "" {
		msg.WriteString(fmt.Sprintf("In-Reply-To: %s\r\n", inReplyTo))
	}
	msg.WriteString("MIME-Version: 1.0\r\n")

	if bodyHTML != "" {
		boundary := "boundary_" + uuid.NewString()[:8]
		msg.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=%s\r\n\r\n", boundary))
		msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
		msg.WriteString(body + "\r\n")
		msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
		msg.WriteString(bodyHTML + "\r\n")
		msg.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else {
		msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
		msg.WriteString(body)
	}

	auth := smtp.PlainAuth("", s.emailCfg.SMTPUser, s.emailCfg.SMTPPass, s.emailCfg.SMTPHost)
	addr := s.emailCfg.SMTPHost + ":" + s.emailCfg.SMTPPort

	if s.emailCfg.SMTPPort == "465" {
		tlsCfg := &tls.Config{ServerName: s.emailCfg.SMTPHost}
		conn, err := tls.Dial("tcp", addr, tlsCfg)
		if err != nil {
			return err
		}
		defer conn.Close()
		client, err := smtp.NewClient(conn, s.emailCfg.SMTPHost)
		if err != nil {
			return err
		}
		if err := client.Auth(auth); err != nil {
			return err
		}
		if err := client.Mail(s.emailCfg.FromEmail); err != nil {
			return err
		}
		for _, r := range to {
			if err := client.Rcpt(r); err != nil {
				return err
			}
		}
		w, err := client.Data()
		if err != nil {
			return err
		}
		_, err = w.Write([]byte(msg.String()))
		w.Close()
		return err
	}

	return smtp.SendMail(addr, auth, s.emailCfg.FromEmail, to, []byte(msg.String()))
}

func (s *communicationService) SendEmail(ctx context.Context, userID string, req dto.SendEmailRequest) (*dto.ThreadResponse, error) {
	now := time.Now()
	msgID := uuid.NewString()

	thread := &model.EmailThread{
		ID:            uuid.NewString(),
		Subject:       req.Subject,
		ContactID:     req.ContactID,
		ContactEmail:  req.To,
		Status:        "open",
		Direction:     "outbound",
		LastMessageAt: now,
		MessageCount:  1,
		UnreadCount:   0,
	}

	if req.ThreadID != "" {
		existing, err := s.repo.FindThreadByID(ctx, req.ThreadID)
		if err == nil && existing != nil {
			msg := &model.EmailMessage{
				ID:        uuid.NewString(),
				ThreadID:  existing.ID,
				MessageID: msgID,
				From:      s.emailCfg.FromEmail,
				To:        req.To,
				Cc:        strings.Join(req.Cc, ", "),
				Subject:   req.Subject,
				Body:      req.Body,
				BodyHTML:  req.BodyHTML,
				Direction: "outbound",
				IsRead:    true,
			}
			sentAt := now
			msg.SentAt = &sentAt

			sendErr := s.sendSMTP([]string{req.To}, req.Subject, req.Body, req.BodyHTML, "", msgID)
			if sendErr != nil {
				slog.Error("SMTP send failed", "error", sendErr)
			}

			if err2 := s.repo.CreateMessage(ctx, msg); err2 != nil {
				return nil, ErrInternal
			}
			existing.MessageCount++
			existing.LastMessageAt = now
			_ = s.repo.UpdateThread(ctx, existing)

			if req.TemplateID != "" {
				_ = s.repo.IncrementTemplateUsage(ctx, req.TemplateID)
			}
			full, _ := s.repo.FindThreadByID(ctx, existing.ID)
			if full == nil {
				full = existing
			}
			r := toThreadResponse(full)
			return &r, nil
		}
	}

	msg := &model.EmailMessage{
		ID:        uuid.NewString(),
		ThreadID:  thread.ID,
		MessageID: msgID,
		From:      s.emailCfg.FromEmail,
		To:        req.To,
		Cc:        strings.Join(req.Cc, ", "),
		Subject:   req.Subject,
		Body:      req.Body,
		BodyHTML:  req.BodyHTML,
		Direction: "outbound",
		IsRead:    true,
	}
	sentAt := now
	msg.SentAt = &sentAt

	sendErr := s.sendSMTP([]string{req.To}, req.Subject, req.Body, req.BodyHTML, "", msgID)
	if sendErr != nil {
		slog.Error("SMTP send failed", "error", sendErr)
	}

	if err := s.repo.CreateThread(ctx, thread); err != nil {
		return nil, ErrInternal
	}
	if err := s.repo.CreateMessage(ctx, msg); err != nil {
		return nil, ErrInternal
	}

	if req.TemplateID != "" {
		_ = s.repo.IncrementTemplateUsage(ctx, req.TemplateID)
	}

	full, _ := s.repo.FindThreadByID(ctx, thread.ID)
	if full == nil {
		thread.Messages = []model.EmailMessage{*msg}
		full = thread
	}
	r := toThreadResponse(full)
	return &r, nil
}

func (s *communicationService) ReplyToThread(ctx context.Context, userID, threadID string, req dto.ReplyEmailRequest) (*dto.MessageResponse, error) {
	thread, err := s.repo.FindThreadByID(ctx, threadID)
	if err != nil || thread == nil {
		return nil, ErrInternal
	}

	now := time.Now()
	msgID := uuid.NewString()

	var inReplyTo string
	if len(thread.Messages) > 0 {
		inReplyTo = thread.Messages[len(thread.Messages)-1].MessageID
	}

	msg := &model.EmailMessage{
		ID:          uuid.NewString(),
		ThreadID:    threadID,
		MessageID:   msgID,
		InReplyTo:   inReplyTo,
		From:        s.emailCfg.FromEmail,
		To:          thread.ContactEmail,
		Cc:          strings.Join(req.Cc, ", "),
		Subject:     "Re: " + thread.Subject,
		Body:        req.Body,
		BodyHTML:    req.BodyHTML,
		Direction:   "outbound",
		IsRead:      true,
	}
	sentAt := now
	msg.SentAt = &sentAt

	sendErr := s.sendSMTP([]string{thread.ContactEmail}, "Re: "+thread.Subject, req.Body, req.BodyHTML, inReplyTo, msgID)
	if sendErr != nil {
		slog.Error("SMTP reply failed", "error", sendErr)
	}

	if err2 := s.repo.CreateMessage(ctx, msg); err2 != nil {
		return nil, ErrInternal
	}

	thread.MessageCount++
	thread.LastMessageAt = now
	_ = s.repo.UpdateThread(ctx, thread)

	r := toMessageResponse(msg)
	return &r, nil
}

func (s *communicationService) ListThreads(ctx context.Context, q dto.ThreadQuery) (*dto.ThreadListResponse, error) {
	pageSize := q.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	page := q.Page
	if page <= 0 {
		page = 1
	}

	threads, total, err := s.repo.ListThreads(ctx, q.Search, q.Status, q.Direction, q.ContactID, page, pageSize)
	if err != nil {
		return nil, ErrInternal
	}

	items := make([]dto.ThreadResponse, 0, len(threads))
	for _, t := range threads {
		items = append(items, toThreadResponse(t))
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	return &dto.ThreadListResponse{
		Data:       items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *communicationService) GetThread(ctx context.Context, id string) (*dto.ThreadResponse, error) {
	t, err := s.repo.FindThreadByID(ctx, id)
	if err != nil {
		return nil, ErrInternal
	}
	if t == nil {
		return nil, nil
	}
	r := toThreadResponse(t)
	return &r, nil
}

func (s *communicationService) DeleteThread(ctx context.Context, id string) error {
	return s.repo.DeleteThread(ctx, id)
}

func (s *communicationService) MarkThreadRead(ctx context.Context, id string) error {
	return s.repo.MarkMessagesRead(ctx, id)
}

func (s *communicationService) ArchiveThread(ctx context.Context, id string) error {
	t, err := s.repo.FindThreadByID(ctx, id)
	if err != nil || t == nil {
		return ErrInternal
	}
	t.Status = "archived"
	return s.repo.UpdateThread(ctx, t)
}

func (s *communicationService) StarMessage(ctx context.Context, msgID string, starred bool) error {
	return s.repo.StarMessage(ctx, msgID, starred)
}

func (s *communicationService) UpdateThreadStatus(ctx context.Context, id, status string) error {
	t, err := s.repo.FindThreadByID(ctx, id)
	if err != nil || t == nil {
		return ErrInternal
	}
	t.Status = status
	return s.repo.UpdateThread(ctx, t)
}

func (s *communicationService) SendBulkEmail(ctx context.Context, userID string, req dto.BulkEmailRequest) (*dto.BulkEmailResponse, error) {
	contactIDsJSON, _ := json.Marshal(req.ContactIDs)

	job := &model.BulkEmailJob{
		ID:          uuid.NewString(),
		Name:        req.Name,
		Subject:     req.Subject,
		Body:        req.Body,
		BodyHTML:    req.BodyHTML,
		TemplateID:  req.TemplateID,
		ContactIDs:  string(contactIDsJSON),
		Status:      "pending",
		TotalCount:  len(req.ContactIDs),
		ScheduledAt: req.ScheduledAt,
		UserID:      userID,
	}

	if err := s.repo.CreateBulkJob(ctx, job); err != nil {
		return nil, ErrInternal
	}

	go s.executeBulkJob(job, req.ContactIDs)

	return toBulkJobResponse(job), nil
}

func (s *communicationService) executeBulkJob(job *model.BulkEmailJob, contactIDs []string) {
	ctx := context.Background()
	now := time.Now()
	job.Status = "running"
	job.StartedAt = &now
	_ = s.repo.UpdateBulkJob(ctx, job)

	for _, cid := range contactIDs {
		thread, err := s.repo.FindThreadByContactAndSubject(ctx, cid, job.Subject)
		if err != nil || thread == nil {
			thread = &model.EmailThread{
				ID:            uuid.NewString(),
				Subject:       job.Subject,
				ContactID:     cid,
				Status:        "open",
				Direction:     "outbound",
				LastMessageAt: time.Now(),
				MessageCount:  1,
			}
			if createErr := s.repo.CreateThread(ctx, thread); createErr != nil {
				job.FailedCount++
				continue
			}
		}

		msgID := uuid.NewString()
		msg := &model.EmailMessage{
			ID:        uuid.NewString(),
			ThreadID:  thread.ID,
			MessageID: msgID,
			From:      s.emailCfg.FromEmail,
			Subject:   job.Subject,
			Body:      job.Body,
			BodyHTML:  job.BodyHTML,
			Direction: "outbound",
			IsRead:    true,
		}
		sentAt := time.Now()
		msg.SentAt = &sentAt

		sendErr := s.sendSMTP([]string{thread.ContactEmail}, job.Subject, job.Body, job.BodyHTML, "", msgID)
		if sendErr != nil {
			slog.Error("Bulk SMTP send failed", "error", sendErr, "contact", cid)
			job.FailedCount++
		} else {
			job.SentCount++
			_ = s.repo.CreateMessage(ctx, msg)
		}

		time.Sleep(100 * time.Millisecond)
	}

	completed := time.Now()
	job.CompletedAt = &completed
	job.Status = "completed"
	_ = s.repo.UpdateBulkJob(ctx, job)

	if job.TemplateID != "" {
		_ = s.repo.IncrementTemplateUsage(ctx, job.TemplateID)
	}
}

func (s *communicationService) ListBulkJobs(ctx context.Context) ([]*dto.BulkEmailResponse, error) {
	jobs, err := s.repo.ListBulkJobs(ctx)
	if err != nil {
		return nil, ErrInternal
	}
	result := make([]*dto.BulkEmailResponse, 0, len(jobs))
	for _, j := range jobs {
		result = append(result, toBulkJobResponse(j))
	}
	return result, nil
}

func (s *communicationService) GetBulkJob(ctx context.Context, id string) (*dto.BulkEmailResponse, error) {
	j, err := s.repo.FindBulkJobByID(ctx, id)
	if err != nil {
		return nil, ErrInternal
	}
	if j == nil {
		return nil, nil
	}
	return toBulkJobResponse(j), nil
}

func (s *communicationService) ListTemplates(ctx context.Context) ([]*dto.TemplateResponse, error) {
	templates, err := s.repo.ListTemplates(ctx)
	if err != nil {
		return nil, ErrInternal
	}
	result := make([]*dto.TemplateResponse, 0, len(templates))
	for _, t := range templates {
		result = append(result, toTemplateResponse(t))
	}
	return result, nil
}

func (s *communicationService) CreateTemplate(ctx context.Context, req dto.TemplateCreateRequest) (*dto.TemplateResponse, error) {
	t := &model.EmailTemplate{
		ID:       uuid.NewString(),
		Name:     req.Name,
		Subject:  req.Subject,
		Body:     req.Body,
		BodyHTML: req.BodyHTML,
		Category: req.Category,
	}
	if err := s.repo.CreateTemplate(ctx, t); err != nil {
		return nil, ErrInternal
	}
	return toTemplateResponse(t), nil
}

func (s *communicationService) UpdateTemplate(ctx context.Context, id string, req dto.TemplateUpdateRequest) (*dto.TemplateResponse, error) {
	t, err := s.repo.FindTemplateByID(ctx, id)
	if err != nil {
		return nil, ErrInternal
	}
	if t == nil {
		return nil, nil
	}
	if req.Name != "" {
		t.Name = req.Name
	}
	if req.Subject != "" {
		t.Subject = req.Subject
	}
	if req.Body != "" {
		t.Body = req.Body
	}
	if req.BodyHTML != "" {
		t.BodyHTML = req.BodyHTML
	}
	if req.Category != "" {
		t.Category = req.Category
	}
	if err := s.repo.UpdateTemplate(ctx, t); err != nil {
		return nil, ErrInternal
	}
	return toTemplateResponse(t), nil
}

func (s *communicationService) DeleteTemplate(ctx context.Context, id string) error {
	return s.repo.DeleteTemplate(ctx, id)
}

func (s *communicationService) ListSignatures(ctx context.Context, userID string) ([]*dto.SignatureResponse, error) {
	sigs, err := s.repo.ListSignatures(ctx, userID)
	if err != nil {
		return nil, ErrInternal
	}
	result := make([]*dto.SignatureResponse, 0, len(sigs))
	for _, sg := range sigs {
		result = append(result, toSignatureResponse(sg))
	}
	return result, nil
}

func (s *communicationService) CreateSignature(ctx context.Context, userID string, req dto.SignatureCreateRequest) (*dto.SignatureResponse, error) {
	if req.IsDefault {
		_ = s.repo.ClearDefaultSignatures(ctx, userID)
	}
	sg := &model.EmailSignature{
		ID:        uuid.NewString(),
		UserID:    userID,
		Name:      req.Name,
		Body:      req.Body,
		IsDefault: req.IsDefault,
	}
	if err := s.repo.CreateSignature(ctx, sg); err != nil {
		return nil, ErrInternal
	}
	return toSignatureResponse(sg), nil
}

func (s *communicationService) UpdateSignature(ctx context.Context, id, userID string, req dto.SignatureUpdateRequest) (*dto.SignatureResponse, error) {
	sg, err := s.repo.FindSignatureByID(ctx, id)
	if err != nil {
		return nil, ErrInternal
	}
	if sg == nil {
		return nil, nil
	}
	if req.IsDefault {
		_ = s.repo.ClearDefaultSignatures(ctx, userID)
	}
	if req.Name != "" {
		sg.Name = req.Name
	}
	if req.Body != "" {
		sg.Body = req.Body
	}
	sg.IsDefault = req.IsDefault
	if err := s.repo.UpdateSignature(ctx, sg); err != nil {
		return nil, ErrInternal
	}
	return toSignatureResponse(sg), nil
}

func (s *communicationService) DeleteSignature(ctx context.Context, id string) error {
	return s.repo.DeleteSignature(ctx, id)
}

func (s *communicationService) GetStats(ctx context.Context) (*dto.StatsResponse, error) {
	stats, err := s.repo.Stats(ctx)
	if err != nil {
		return nil, ErrInternal
	}

	recentThreads, _ := s.repo.RecentContacts(ctx, 10)
	recentMap := make(map[string]*dto.RecentContactResponse)
	for _, t := range recentThreads {
		if t.ContactID == "" {
			continue
		}
		if existing, ok := recentMap[t.ContactID]; ok {
			existing.ThreadCount++
			if t.LastMessageAt.After(existing.LastContact) {
				existing.LastContact = t.LastMessageAt
			}
		} else {
			recentMap[t.ContactID] = &dto.RecentContactResponse{
				ContactID:    t.ContactID,
				ContactName:  t.ContactName,
				ContactEmail: t.ContactEmail,
				LastContact:  t.LastMessageAt,
				ThreadCount:  1,
			}
		}
	}

	recentList := make([]dto.RecentContactResponse, 0, len(recentMap))
	for _, r := range recentMap {
		recentList = append(recentList, *r)
	}
	if len(recentList) > 8 {
		recentList = recentList[:8]
	}

	return &dto.StatsResponse{
		TotalThreads:   stats["total_threads"],
		OpenThreads:    stats["open_threads"],
		UnreadMessages: stats["unread_messages"],
		SentToday:      stats["sent_today"],
		BulkJobsRun:    stats["bulk_jobs_run"],
		RecentContacts: recentList,
	}, nil
}

func (s *communicationService) SyncInbox(ctx context.Context) (*dto.SyncStatusResponse, error) {
	cfg := s.emailCfg
	if cfg.SMTPUser == "" {
		return &dto.SyncStatusResponse{
			Connected:    false,
			EmailAddress: "",
			Error:        "SMTP/IMAP not configured. Set SMTP_USER, SMTP_PASS, SMTP_HOST env vars.",
		}, nil
	}

	now := time.Now()
	return &dto.SyncStatusResponse{
		Connected:    true,
		LastSync:     &now,
		EmailAddress: cfg.SMTPUser,
	}, nil
}

func (s *communicationService) GetContactThreads(ctx context.Context, contactID string) ([]*dto.ThreadResponse, error) {
	threads, _, err := s.repo.ListThreads(ctx, "", "", "", contactID, 1, 50)
	if err != nil {
		return nil, ErrInternal
	}
	result := make([]*dto.ThreadResponse, 0, len(threads))
	for _, t := range threads {
		r := toThreadResponse(t)
		result = append(result, &r)
	}
	return result, nil
}
