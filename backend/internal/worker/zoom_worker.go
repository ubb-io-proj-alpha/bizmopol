package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"backend/internal/repository"
	"backend/internal/ws"
	"backend/internal/zoom"
)

type ZoomJobType string

const (
	ZoomJobSync   ZoomJobType = "sync"
	ZoomJobDelete ZoomJobType = "delete"
	ZoomJobCreate ZoomJobType = "create"
)

type ZoomJob struct {
	Type      ZoomJobType
	UserID    string
	EventID   string
	MeetingID int64
	Title     string
	Desc      string
	StartTime time.Time
	EndTime   time.Time
}

type ZoomWorker struct {
	jobs chan ZoomJob
	repo repository.CalendarRepository
	hub  *ws.Hub
	done chan struct{}
}

func NewZoomWorker(repo repository.CalendarRepository, hub *ws.Hub) *ZoomWorker {
	return &ZoomWorker{
		jobs: make(chan ZoomJob, 64),
		repo: repo,
		hub:  hub,
		done: make(chan struct{}),
	}
}

func (w *ZoomWorker) Enqueue(job ZoomJob) {
	select {
	case w.jobs <- job:
		slog.Info("ZoomWorker: job enqueued", "type", job.Type, "event_id", job.EventID, "zoom_meeting_id", job.MeetingID)
	default:
		slog.Error("ZoomWorker: queue full, dropping job", "type", job.Type, "event_id", job.EventID)
	}
}

func (w *ZoomWorker) Start(ctx context.Context) {
	slog.Info("ZoomWorker: started")
	go func() {
		defer close(w.done)
		for {
			select {
			case job := <-w.jobs:
				w.process(ctx, job)
			case <-ctx.Done():
				slog.Info("ZoomWorker: shutting down, draining remaining jobs")
				for {
					select {
					case job := <-w.jobs:
						w.process(context.Background(), job)
					default:
						slog.Info("ZoomWorker: shutdown complete")
						return
					}
				}
			}
		}
	}()
}

func (w *ZoomWorker) Wait() {
	<-w.done
}

func (w *ZoomWorker) process(ctx context.Context, job ZoomJob) {
	slog.Info("ZoomWorker: processing job", "type", job.Type, "event_id", job.EventID, "zoom_meeting_id", job.MeetingID)
	start := time.Now()

	var err error
	switch job.Type {
	case ZoomJobSync:
		err = w.processSync(ctx, job)
	case ZoomJobDelete:
		err = w.processDelete(ctx, job)
	case ZoomJobCreate:
		err = w.processCreate(ctx, job)
	default:
		slog.Error("ZoomWorker: unknown job type", "type", job.Type)
		return
	}

	if err != nil {
		slog.Error("ZoomWorker: job failed", "type", job.Type, "event_id", job.EventID, "error", err, "duration", time.Since(start))
		w.notify(job, false, err.Error())
	} else {
		slog.Info("ZoomWorker: job completed", "type", job.Type, "event_id", job.EventID, "duration", time.Since(start))
		w.notify(job, true, "")
	}
}

func (w *ZoomWorker) notify(job ZoomJob, success bool, errMsg string) {
	if w.hub == nil {
		return
	}

	var nType, message string
	data := map[string]interface{}{
		"event_id": job.EventID,
		"job_type": string(job.Type),
	}

	if success {
		nType = "zoom_job_done"
		switch job.Type {
		case ZoomJobCreate:
			message = "Spotkanie Zoom zostało utworzone"
		case ZoomJobDelete:
			message = "Spotkanie Zoom zostało usunięte"
		case ZoomJobSync:
			message = "Spotkanie Zoom zostało zsynchronizowane"
		}
	} else {
		nType = "zoom_job_failed"
		data["error"] = errMsg
		switch job.Type {
		case ZoomJobCreate:
			message = "Nie udało się utworzyć spotkania Zoom"
		case ZoomJobDelete:
			message = "Nie udało się usunąć spotkania Zoom"
		case ZoomJobSync:
			message = "Nie udało się zsynchronizować spotkania Zoom"
		}
	}

	w.hub.SendToUser(job.UserID, ws.Notification{
		Type:    nType,
		Message: message,
		Data:    data,
	})
}

func (w *ZoomWorker) getToken(ctx context.Context, userID string) (string, error) {
	account, err := w.repo.FindZoomAccountByUserID(ctx, userID)
	if err != nil || account == nil {
		return "", fmt.Errorf("zoom not configured for user %s", userID)
	}
	return zoom.GetAccessToken(account.AccountID, account.ClientID, account.ClientSecret)
}

func (w *ZoomWorker) processSync(ctx context.Context, job ZoomJob) error {
	token, err := w.getToken(ctx, job.UserID)
	if err != nil {
		return fmt.Errorf("get token: %w", err)
	}

	duration := int(job.EndTime.Sub(job.StartTime).Minutes())
	if duration <= 0 {
		duration = 30
	}

	slog.Info("ZoomWorker: syncing meeting", "meeting_id", job.MeetingID, "title", job.Title, "start", job.StartTime, "duration", duration)

	if err := zoom.UpdateMeeting(token, job.MeetingID, job.Title, job.Desc, job.StartTime, duration); err != nil {
		return fmt.Errorf("update zoom meeting: %w", err)
	}

	return nil
}

func (w *ZoomWorker) processDelete(ctx context.Context, job ZoomJob) error {
	token, err := w.getToken(ctx, job.UserID)
	if err != nil {
		return fmt.Errorf("get token: %w", err)
	}

	slog.Info("ZoomWorker: deleting meeting", "meeting_id", job.MeetingID)

	if err := zoom.DeleteMeeting(token, job.MeetingID); err != nil {
		return fmt.Errorf("delete zoom meeting: %w", err)
	}

	return nil
}

func (w *ZoomWorker) processCreate(ctx context.Context, job ZoomJob) error {
	token, err := w.getToken(ctx, job.UserID)
	if err != nil {
		return fmt.Errorf("get token: %w", err)
	}

	duration := int(job.EndTime.Sub(job.StartTime).Minutes())
	if duration <= 0 {
		duration = 30
	}

	slog.Info("ZoomWorker: creating meeting", "event_id", job.EventID, "title", job.Title, "start", job.StartTime, "duration", duration)

	meeting, err := zoom.CreateMeeting(token, job.Title, job.Desc, job.StartTime, duration)
	if err != nil {
		return fmt.Errorf("create zoom meeting: %w", err)
	}

	event, err := w.repo.FindEventByID(ctx, job.EventID)
	if err != nil || event == nil {
		return fmt.Errorf("event %s not found after zoom creation", job.EventID)
	}

	event.ZoomMeetingID = meeting.ID
	event.ZoomJoinURL = meeting.JoinURL
	event.ZoomStartURL = meeting.StartURL
	event.ZoomPasscode = meeting.Password
	if event.Location == "" {
		event.Location = meeting.JoinURL
	}

	if err := w.repo.UpdateEvent(ctx, event); err != nil {
		return fmt.Errorf("update event with zoom data: %w", err)
	}

	slog.Info("ZoomWorker: event updated with zoom", "event_id", job.EventID, "zoom_meeting_id", meeting.ID, "join_url", meeting.JoinURL)
	return nil
}
