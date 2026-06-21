package worker

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/ws"
)

type EmailJobType string

const (
	EmailJobSend EmailJobType = "send"
	EmailJobSync EmailJobType = "sync"
)

type EmailJob struct {
	Type      EmailJobType
	UserID    string
	MessageID string
	ThreadID  string
	To        []string
	Subject   string
	Body      string
	BodyHTML  string
	InReplyTo string
	SmtpMsgID string

	BulkJobID    string
	BulkSent     int
	BulkTotal    int
	ContactID    string
	ContactEmail string
	ContactName  string
	FromEmail    string
	DeferRecord  bool
}

type bulkProgress struct {
	sent   int
	failed int
	total  int
}

type SyncFunc func(ctx context.Context, userID string) (newMessages int, err error)

type EmailWorker struct {
	jobs     chan EmailJob
	repo     repository.CommunicationRepository
	hub      *ws.Hub
	done     chan struct{}
	sendFunc func(to []string, subject, body, bodyHTML, inReplyTo, msgID, smtpHost, smtpPort, smtpUser, smtpPass, fromEmail, fromName string) error
	syncFunc SyncFunc

	bulkMu    sync.Mutex
	bulkState map[string]*bulkProgress
}

func NewEmailWorker(repo repository.CommunicationRepository, hub *ws.Hub) *EmailWorker {
	return &EmailWorker{
		jobs:      make(chan EmailJob, 128),
		repo:      repo,
		hub:       hub,
		done:      make(chan struct{}),
		bulkState: make(map[string]*bulkProgress),
	}
}

func (w *EmailWorker) SetSendFunc(fn func(to []string, subject, body, bodyHTML, inReplyTo, msgID, smtpHost, smtpPort, smtpUser, smtpPass, fromEmail, fromName string) error) {
	w.sendFunc = fn
}

func (w *EmailWorker) SetSyncFunc(fn SyncFunc) {
	w.syncFunc = fn
}

func (w *EmailWorker) Enqueue(job EmailJob) {
	select {
	case w.jobs <- job:
		slog.Info("EmailWorker: job enqueued", "type", job.Type, "user_id", job.UserID, "message_id", job.MessageID)
	default:
		slog.Error("EmailWorker: queue full, dropping job", "type", job.Type, "user_id", job.UserID)
	}
}

func (w *EmailWorker) QueueLen() int {
	return len(w.jobs)
}

func (w *EmailWorker) Notify(userID string, n ws.Notification) {
	if w.hub != nil {
		w.hub.SendToUser(userID, n)
	}
}

func (w *EmailWorker) Start(ctx context.Context) {
	slog.Info("EmailWorker: started")
	go func() {
		defer close(w.done)
		for {
			select {
			case job := <-w.jobs:
				w.process(ctx, job)
			case <-ctx.Done():
				slog.Info("EmailWorker: shutting down, draining remaining jobs")
				for {
					select {
					case job := <-w.jobs:
						w.process(context.Background(), job)
					default:
						slog.Info("EmailWorker: shutdown complete")
						return
					}
				}
			}
		}
	}()

	go w.imapPollLoop(ctx)
}

func (w *EmailWorker) imapPollLoop(ctx context.Context) {
	time.Sleep(5 * time.Second)
	slog.Info("EmailWorker: IMAP poll loop started")
	for {
		delay := 10 + rand.Intn(6)
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(delay) * time.Second):
		}

		if w.syncFunc == nil {
			continue
		}

		accounts, err := w.repo.ListActiveEmailAccounts(ctx)
		if err != nil {
			slog.Error("EmailWorker: failed to list accounts for IMAP poll", "error", err)
			continue
		}

		for _, acc := range accounts {
			if acc.IMAPHost == "" {
				continue
			}
			newMsgs, err := w.syncFunc(ctx, acc.UserID)
			if err != nil {
				slog.Error("EmailWorker: IMAP poll sync failed", "user_id", acc.UserID, "error", err)
				continue
			}
			if newMsgs > 0 {
				slog.Info("EmailWorker: IMAP poll found new messages", "user_id", acc.UserID, "count", newMsgs)
			}
		}
	}
}

func (w *EmailWorker) Wait() {
	<-w.done
}

func (w *EmailWorker) process(ctx context.Context, job EmailJob) {
	slog.Info("EmailWorker: processing", "type", job.Type, "user_id", job.UserID)
	start := time.Now()

	var err error
	switch job.Type {
	case EmailJobSend:
		if job.BulkJobID != "" {
			time.Sleep(1 * time.Second)
		}
		err = w.processSend(ctx, job)
	case EmailJobSync:
		return
	default:
		slog.Error("EmailWorker: unknown job type", "type", job.Type)
		return
	}

	if err != nil {
		slog.Error("EmailWorker: job failed", "type", job.Type, "user_id", job.UserID, "error", err, "duration", time.Since(start))
	} else {
		slog.Info("EmailWorker: job done", "type", job.Type, "user_id", job.UserID, "duration", time.Since(start))
	}

	if job.BulkJobID != "" {
		w.trackBulkProgress(ctx, job, err == nil)
	} else {
		remaining := len(w.jobs)
		w.notify(job, err == nil, errStr(err), remaining)
	}
}

func (w *EmailWorker) processSend(ctx context.Context, job EmailJob) error {
	if w.sendFunc == nil {
		return fmt.Errorf("send function not configured")
	}

	account, err := w.repo.FindEmailAccountByUserID(ctx, job.UserID)
	if err != nil || account == nil {
		return fmt.Errorf("no email account for user %s", job.UserID)
	}

	smtpHost := account.SMTPHost
	smtpPort := fmt.Sprintf("%d", account.SMTPPort)
	smtpUser := account.Email
	smtpPass := account.Password
	fromEmail := account.Email
	fromName := "BizmoPol CRM"

	slog.Info("EmailWorker: sending via SMTP", "to", job.To, "subject", job.Subject, "smtp_host", smtpHost)

	if err := w.sendFunc(job.To, job.Subject, job.Body, job.BodyHTML, job.InReplyTo, job.SmtpMsgID, smtpHost, smtpPort, smtpUser, smtpPass, fromEmail, fromName); err != nil {
		return fmt.Errorf("SMTP send: %w", err)
	}

	if job.DeferRecord {
		w.createDeferredRecord(ctx, job)
	}

	return nil
}

func (w *EmailWorker) createDeferredRecord(ctx context.Context, job EmailJob) {
	thread, _ := w.repo.FindThreadByContactAndSubject(ctx, job.ContactID, job.Subject)
	if thread == nil {
		thread = &model.EmailThread{
			ID:            job.ThreadID,
			Subject:       job.Subject,
			ContactID:     job.ContactID,
			ContactEmail:  job.ContactEmail,
			ContactName:   job.ContactName,
			Status:        "open",
			Direction:     "outbound",
			LastMessageAt: time.Now(),
			MessageCount:  1,
		}
		if err := w.repo.CreateThread(ctx, thread); err != nil {
			slog.Error("EmailWorker: failed to create thread", "error", err)
			return
		}
	} else {
		thread.MessageCount++
		thread.LastMessageAt = time.Now()
		_ = w.repo.UpdateThread(ctx, thread)
	}

	now := time.Now()
	msg := &model.EmailMessage{
		ID:        job.MessageID,
		ThreadID:  thread.ID,
		MessageID: job.SmtpMsgID,
		From:      job.FromEmail,
		To:        job.ContactEmail,
		Subject:   job.Subject,
		Body:      job.Body,
		BodyHTML:  job.BodyHTML,
		Direction: "outbound",
		IsRead:    true,
		SentAt:    &now,
	}
	if err := w.repo.CreateMessage(ctx, msg); err != nil {
		slog.Error("EmailWorker: failed to create message", "error", err)
	}
}

func errStr(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}

func (w *EmailWorker) trackBulkProgress(ctx context.Context, job EmailJob, success bool) {
	w.bulkMu.Lock()
	bp, ok := w.bulkState[job.BulkJobID]
	if !ok {
		bp = &bulkProgress{total: job.BulkTotal}
		w.bulkState[job.BulkJobID] = bp
	}
	if success {
		bp.sent++
	} else {
		bp.failed++
	}
	sent := bp.sent
	failed := bp.failed
	total := bp.total
	done := (sent + failed) >= total
	if done {
		delete(w.bulkState, job.BulkJobID)
	}
	w.bulkMu.Unlock()

	data := map[string]interface{}{
		"bulk_job_id": job.BulkJobID,
		"bulk_sent":   sent,
		"bulk_failed": failed,
		"bulk_total":  total,
	}

	if done {
		bulkJob, _ := w.repo.FindBulkJobByID(ctx, job.BulkJobID)
		if bulkJob != nil {
			now := time.Now()
			bulkJob.Status = "completed"
			bulkJob.SentCount = sent
			bulkJob.FailedCount = failed
			bulkJob.CompletedAt = &now
			_ = w.repo.UpdateBulkJob(ctx, bulkJob)
		}
		data["completed"] = true
		w.hub.SendToUser(job.UserID, ws.Notification{
			Type:    "bulk_completed",
			Message: fmt.Sprintf("Kampania zakończona - %d wysłanych", sent),
			Data:    data,
		})
	} else {
		w.hub.SendToUser(job.UserID, ws.Notification{
			Type: "bulk_progress",
			Data: data,
		})
	}
}

func (w *EmailWorker) notify(job EmailJob, success bool, errMsg string, queueRemaining int) {
	if w.hub == nil {
		return
	}

	data := map[string]interface{}{
		"job_type":        string(job.Type),
		"queue_remaining": queueRemaining,
	}

	var nType, message string

	if job.Type == EmailJobSend {
		data["thread_id"] = job.ThreadID
		data["message_id"] = job.MessageID
		if job.BulkJobID != "" {
			data["bulk_job_id"] = job.BulkJobID
			data["bulk_sent"] = job.BulkSent
			data["bulk_total"] = job.BulkTotal
		}
		if success {
			nType = "email_sent"
			message = "Wiadomość została wysłana"
		} else {
			nType = "email_send_failed"
			message = "Nie udało się wysłać wiadomości"
			data["error"] = errMsg
		}
	} else if job.Type == EmailJobSync {
		if success {
			nType = "email_sync_done"
			message = "Synchronizacja zakończona"
		} else {
			nType = "email_sync_failed"
			message = "Błąd synchronizacji"
			data["error"] = errMsg
		}
	}

	w.hub.SendToUser(job.UserID, ws.Notification{
		Type:    nType,
		Message: message,
		Data:    data,
	})
}
