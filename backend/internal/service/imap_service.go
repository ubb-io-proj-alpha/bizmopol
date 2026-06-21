package service

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"backend/internal/model"
	"backend/internal/repository"
)

type IMAPSyncResult struct {
	NewMessages int
	Error       string
}

type IMAPClient struct {
	conn   net.Conn
	reader *bufio.Reader
	tag    int
}

func newIMAPClient(host string, port int) (*IMAPClient, error) {
	addr := fmt.Sprintf("%s:%d", host, port)

	var conn net.Conn
	var err error

	if port == 993 {
		tlsConfig := &tls.Config{ServerName: host}
		conn, err = tls.DialWithDialer(&net.Dialer{Timeout: 15 * time.Second}, "tcp", addr, tlsConfig)
	} else {
		conn, err = net.DialTimeout("tcp", addr, 15*time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("IMAP connect failed: %w", err)
	}

	client := &IMAPClient{
		conn:   conn,
		reader: bufio.NewReader(conn),
		tag:    0,
	}

	if _, err := client.readResponse("*"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("IMAP greeting failed: %w", err)
	}

	return client, nil
}

func (c *IMAPClient) nextTag() string {
	c.tag++
	return fmt.Sprintf("A%03d", c.tag)
}

func (c *IMAPClient) sendCommand(command string) (string, error) {
	tag := c.nextTag()
	line := fmt.Sprintf("%s %s\r\n", tag, command)
	if _, err := c.conn.Write([]byte(line)); err != nil {
		return "", err
	}
	return tag, nil
}

func (c *IMAPClient) readResponse(tag string) ([]string, error) {
	var lines []string
	c.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	for {
		line, err := c.reader.ReadString('\n')
		if err != nil {
			return lines, err
		}
		line = strings.TrimRight(line, "\r\n")
		lines = append(lines, line)

		if strings.HasPrefix(line, tag+" OK") {
			return lines, nil
		}
		if strings.HasPrefix(line, tag+" NO") || strings.HasPrefix(line, tag+" BAD") {
			return lines, fmt.Errorf("IMAP error: %s", line)
		}
	}
}

func (c *IMAPClient) readLiteral(size int) (string, error) {
	buf := make([]byte, size)
	c.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	_, err := io.ReadFull(c.reader, buf)
	if err != nil {
		return "", err
	}
	c.reader.ReadString('\n')
	return string(buf), nil
}

func (c *IMAPClient) login(user, pass string) error {
	// Try AUTHENTICATE PLAIN first (required by Outlook/Exchange)
	plain := fmt.Sprintf("\x00%s\x00%s", user, pass)
	encoded := base64.StdEncoding.EncodeToString([]byte(plain))
	tag, err := c.sendCommand(fmt.Sprintf("AUTHENTICATE PLAIN %s", encoded))
	if err == nil {
		if _, err2 := c.readResponse(tag); err2 == nil {
			slog.Info("IMAP auth OK via AUTHENTICATE PLAIN")
			return nil
		}
	}

	// Fall back to LOGIN
	tag, err = c.sendCommand(fmt.Sprintf("LOGIN %s %s", quoteIMAPString(user), quoteIMAPString(pass)))
	if err != nil {
		return err
	}
	_, err = c.readResponse(tag)
	return err
}

func (c *IMAPClient) selectMailbox(name string) (int, error) {
	tag, err := c.sendCommand(fmt.Sprintf("SELECT %s", quoteIMAPString(name)))
	if err != nil {
		return 0, err
	}
	lines, err := c.readResponse(tag)
	if err != nil {
		return 0, err
	}
	exists := 0
	for _, line := range lines {
		if strings.Contains(line, "EXISTS") {
			fmt.Sscanf(line, "* %d EXISTS", &exists)
		}
	}
	return exists, nil
}

func (c *IMAPClient) searchSince(since time.Time) ([]int, error) {
	dateStr := since.Format("02-Jan-2006")
	tag, err := c.sendCommand(fmt.Sprintf("SEARCH SINCE %s", dateStr))
	if err != nil {
		return nil, err
	}
	lines, err := c.readResponse(tag)
	if err != nil {
		return nil, err
	}
	var uids []int
	for _, line := range lines {
		if strings.HasPrefix(line, "* SEARCH") {
			parts := strings.Fields(line)
			for _, p := range parts[2:] {
				if n, err := strconv.Atoi(p); err == nil {
					uids = append(uids, n)
				}
			}
		}
	}
	return uids, nil
}

func (c *IMAPClient) searchUID(sinceUID uint32) ([]int, error) {
	tag, err := c.sendCommand(fmt.Sprintf("UID SEARCH UID %d:*", sinceUID+1))
	if err != nil {
		return nil, err
	}
	lines, err := c.readResponse(tag)
	if err != nil {
		return nil, err
	}
	var uids []int
	for _, line := range lines {
		if strings.HasPrefix(line, "* SEARCH") {
			parts := strings.Fields(line)
			for _, p := range parts[2:] {
				if n, err := strconv.Atoi(p); err == nil {
					if uint32(n) > sinceUID {
						uids = append(uids, n)
					}
				}
			}
		}
	}
	return uids, nil
}

type fetchedEmail struct {
	UID       uint32
	MessageID string
	InReplyTo string
	From      string
	To        string
	Cc        string
	Subject   string
	Date      time.Time
	Body      string
	BodyHTML  string
}

func (c *IMAPClient) fetchMessages(seqNums []int, useUID bool) ([]fetchedEmail, error) {
	if len(seqNums) == 0 {
		return nil, nil
	}

	var results []fetchedEmail

	batchSize := 25
	for i := 0; i < len(seqNums); i += batchSize {
		end := i + batchSize
		if end > len(seqNums) {
			end = len(seqNums)
		}
		batch := seqNums[i:end]

		seqStr := intSliceToRange(batch)
		var cmd string
		if useUID {
			cmd = fmt.Sprintf("UID FETCH %s (UID FLAGS BODY[HEADER.FIELDS (FROM TO CC SUBJECT DATE MESSAGE-ID IN-REPLY-TO)] BODY[TEXT])", seqStr)
		} else {
			cmd = fmt.Sprintf("FETCH %s (UID FLAGS BODY[HEADER.FIELDS (FROM TO CC SUBJECT DATE MESSAGE-ID IN-REPLY-TO)] BODY[TEXT])", seqStr)
		}

		tag, err := c.sendCommand(cmd)
		if err != nil {
			return results, err
		}

		emails, err := c.parseFetchResponse(tag)
		if err != nil {
			return results, err
		}
		results = append(results, emails...)
	}

	return results, nil
}

func (c *IMAPClient) parseFetchResponse(tag string) ([]fetchedEmail, error) {
	var results []fetchedEmail
	var currentEmail *fetchedEmail
	literalRe := regexp.MustCompile(`\{(\d+)\}$`)

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	for {
		line, err := c.reader.ReadString('\n')
		if err != nil {
			return results, err
		}
		line = strings.TrimRight(line, "\r\n")

		if strings.HasPrefix(line, tag+" OK") {
			if currentEmail != nil {
				results = append(results, *currentEmail)
			}
			return results, nil
		}
		if strings.HasPrefix(line, tag+" NO") || strings.HasPrefix(line, tag+" BAD") {
			return results, fmt.Errorf("IMAP FETCH error: %s", line)
		}

		if strings.HasPrefix(line, "* ") && strings.Contains(line, "FETCH") {
			if currentEmail != nil {
				results = append(results, *currentEmail)
			}
			currentEmail = &fetchedEmail{Date: time.Now()}

			if uidMatch := regexp.MustCompile(`UID (\d+)`).FindStringSubmatch(line); len(uidMatch) > 1 {
				if n, err := strconv.ParseUint(uidMatch[1], 10, 32); err == nil {
					currentEmail.UID = uint32(n)
				}
			}
		}

		if matches := literalRe.FindStringSubmatch(line); len(matches) > 1 {
			size, _ := strconv.Atoi(matches[1])
			data, err := c.readLiteral(size)
			if err != nil {
				return results, err
			}
			if currentEmail != nil {
				if strings.Contains(line, "HEADER.FIELDS") {
					parseHeaders(currentEmail, data)
				} else if strings.Contains(line, "BODY[TEXT]") || strings.Contains(line, "BODY[1]") {
					currentEmail.Body = cleanBody(data)
				}
			}
		}
	}
}

func parseHeaders(email *fetchedEmail, headers string) {
	lines := strings.Split(headers, "\n")
	var currentKey, currentVal string
	flush := func() {
		if currentKey == "" {
			return
		}
		val := strings.TrimSpace(currentVal)
		switch strings.ToLower(currentKey) {
		case "from":
			email.From = decodeHeader(val)
		case "to":
			email.To = decodeHeader(val)
		case "cc":
			email.Cc = decodeHeader(val)
		case "subject":
			email.Subject = decodeHeader(val)
		case "date":
			email.Date = parseDate(val)
		case "message-id":
			email.MessageID = strings.Trim(val, "<> ")
		case "in-reply-to":
			email.InReplyTo = strings.Trim(val, "<> ")
		}
	}
	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		if line[0] == ' ' || line[0] == '\t' {
			currentVal += " " + strings.TrimSpace(line)
			continue
		}
		flush()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			currentKey = strings.TrimSpace(parts[0])
			currentVal = strings.TrimSpace(parts[1])
		}
	}
	flush()
}

func (c *IMAPClient) logout() {
	tag, _ := c.sendCommand("LOGOUT")
	c.readResponse(tag)
	c.conn.Close()
}

func SyncIMAPInbox(ctx context.Context, account *model.EmailAccount, repo repository.CommunicationRepository) IMAPSyncResult {
	if account.IMAPHost == "" || account.Email == "" || account.Password == "" {
		return IMAPSyncResult{Error: "IMAP not configured. Provide host, email, and password."}
	}

	client, err := newIMAPClient(account.IMAPHost, account.IMAPPort)
	if err != nil {
		slog.Error("IMAP connection failed", "error", err, "host", account.IMAPHost)
		return IMAPSyncResult{Error: fmt.Sprintf("Connection failed: %v", err)}
	}
	defer client.logout()

	if err := client.login(account.Email, account.Password); err != nil {
		slog.Error("IMAP login failed", "error", err, "email", account.Email)
		return IMAPSyncResult{Error: "Authentication failed. Check email and app password."}
	}

	_, err = client.selectMailbox("INBOX")
	if err != nil {
		return IMAPSyncResult{Error: fmt.Sprintf("Failed to select INBOX: %v", err)}
	}

	var seqNums []int
	useUID := false
	if account.LastSyncUID > 0 {
		seqNums, err = client.searchUID(account.LastSyncUID)
		useUID = true
	} else {
		since := time.Now().AddDate(0, 0, -30)
		seqNums, err = client.searchSince(since)
	}
	if err != nil {
		return IMAPSyncResult{Error: fmt.Sprintf("Search failed: %v", err)}
	}

	if len(seqNums) == 0 {
		now := time.Now()
		account.LastSyncAt = &now
		repo.UpdateEmailAccount(ctx, account)
		return IMAPSyncResult{NewMessages: 0}
	}

	slog.Info("IMAP sync found messages", "count", len(seqNums))

	emails, err := client.fetchMessages(seqNums, useUID)
	if err != nil {
		slog.Error("IMAP fetch failed", "error", err)
		return IMAPSyncResult{Error: fmt.Sprintf("Fetch failed: %v", err)}
	}

	newCount := 0
	var maxUID uint32

	for _, email := range emails {
		if email.UID > maxUID {
			maxUID = email.UID
		}

		msgID := email.MessageID
		if msgID == "" {
			msgID = fmt.Sprintf("imap-%s-%d", account.Email, email.UID)
		}

		existing, _ := repo.FindMessageByMessageID(ctx, msgID)
		if existing != nil {
			continue
		}

		fromAddr := extractEmail(email.From)
		toAddr := extractEmail(email.To)

		isInbound := !strings.EqualFold(fromAddr, account.Email)

		var thread *model.EmailThread

		if email.InReplyTo != "" {
			parentMsg, _ := repo.FindMessageByMessageID(ctx, strings.Trim(email.InReplyTo, "<> "))
			if parentMsg != nil {
				t, _ := repo.FindThreadByID(ctx, parentMsg.ThreadID)
				if t != nil {
					thread = t
				}
			}
		}

		if thread == nil {
			subject := normalizeSubject(email.Subject)
			contactEmail := fromAddr
			if !isInbound {
				contactEmail = toAddr
			}
			threads, _, _ := repo.ListThreads(ctx, subject, "", "", "", 1, 5)
			for _, t := range threads {
				if strings.EqualFold(t.ContactEmail, contactEmail) {
					thread = t
					break
				}
			}
		}

		if thread == nil {
			contactEmail := fromAddr
			contactName := extractName(email.From)
			direction := "inbound"
			if !isInbound {
				contactEmail = toAddr
				contactName = extractName(email.To)
				direction = "outbound"
			}
			thread = &model.EmailThread{
				ID:            uuid.NewString(),
				Subject:       email.Subject,
				ContactEmail:  contactEmail,
				ContactName:   contactName,
				Status:        "open",
				Direction:     direction,
				LastMessageAt: email.Date,
				MessageCount:  0,
				UnreadCount:   0,
			}
			if err := repo.CreateThread(ctx, thread); err != nil {
				slog.Error("Failed to create thread from IMAP", "error", err)
				continue
			}
		}

		direction := "inbound"
		if !isInbound {
			direction = "outbound"
		}

		sentAt := email.Date
		msg := &model.EmailMessage{
			ID:        uuid.NewString(),
			ThreadID:  thread.ID,
			MessageID: msgID,
			InReplyTo: email.InReplyTo,
			From:      email.From,
			To:        email.To,
			Cc:        email.Cc,
			Subject:   email.Subject,
			Body:      email.Body,
			Direction: direction,
			IsRead:    !isInbound,
			SentAt:    &sentAt,
		}

		if err := repo.CreateMessage(ctx, msg); err != nil {
			slog.Error("Failed to save IMAP message", "error", err, "messageID", msgID)
			continue
		}

		thread.MessageCount++
		if email.Date.After(thread.LastMessageAt) {
			thread.LastMessageAt = email.Date
		}
		if isInbound {
			thread.UnreadCount++
		}
		repo.UpdateThread(ctx, thread)

		newCount++
	}

	now := time.Now()
	account.LastSyncAt = &now
	if maxUID > account.LastSyncUID {
		account.LastSyncUID = maxUID
	}
	account.SyncedCount += newCount
	repo.UpdateEmailAccount(ctx, account)

	slog.Info("IMAP sync completed", "newMessages", newCount)
	return IMAPSyncResult{NewMessages: newCount}
}

func TestIMAPConnection(host string, port int, email, password string) error {
	slog.Info("Testing IMAP connection", "host", host, "port", port, "email", email)
	client, err := newIMAPClient(host, port)
	if err != nil {
		slog.Error("IMAP connection test failed", "error", err)
		return err
	}
	defer client.logout()

	if port == 1993 {
		slog.Info("IMAP connection test OK (mailpit, no auth required)")
		return nil
	}

	if err := client.login(email, password); err != nil {
		slog.Error("IMAP auth test failed", "error", err)
		return fmt.Errorf("authentication failed: %w", err)
	}
	slog.Info("IMAP connection test OK")
	return nil
}

func TestSMTPConnection(host string, port int, email, password string) error {
	slog.Info("Testing SMTP connection", "host", host, "port", port, "email", email)
	addr := fmt.Sprintf("%s:%d", host, port)
	if port == 465 {
		tlsCfg := &tls.Config{ServerName: host}
		conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 10 * time.Second}, "tcp", addr, tlsCfg)
		if err != nil {
			slog.Error("SMTP TLS connection test failed", "error", err)
			return fmt.Errorf("SMTP TLS connection failed: %w", err)
		}
		conn.Close()
		slog.Info("SMTP connection test OK (TLS)")
		return nil
	}
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		slog.Error("SMTP connection test failed", "error", err)
		return fmt.Errorf("SMTP connection failed: %w", err)
	}
	conn.Close()
	slog.Info("SMTP connection test OK")
	return nil
}

func quoteIMAPString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `"` + s + `"`
}

func intSliceToRange(nums []int) string {
	parts := make([]string, len(nums))
	for i, n := range nums {
		parts[i] = strconv.Itoa(n)
	}
	return strings.Join(parts, ",")
}

func cleanBody(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 50000 {
		s = s[:50000]
	}
	return s
}

func normalizeSubject(s string) string {
	s = strings.TrimSpace(s)
	for {
		lower := strings.ToLower(s)
		if strings.HasPrefix(lower, "re:") || strings.HasPrefix(lower, "fw:") {
			s = strings.TrimSpace(s[3:])
			continue
		}
		if strings.HasPrefix(lower, "fwd:") {
			s = strings.TrimSpace(s[4:])
			continue
		}
		break
	}
	return s
}

func extractEmail(addr string) string {
	if idx := strings.Index(addr, "<"); idx >= 0 {
		end := strings.Index(addr, ">")
		if end > idx {
			return strings.TrimSpace(addr[idx+1 : end])
		}
	}
	return strings.TrimSpace(addr)
}

func extractName(addr string) string {
	if idx := strings.Index(addr, "<"); idx > 0 {
		name := strings.TrimSpace(addr[:idx])
		name = strings.Trim(name, `"'`)
		if name != "" {
			return name
		}
	}
	return extractEmail(addr)
}

func parseDate(s string) time.Time {
	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		"Mon, 2 Jan 2006 15:04:05 -0700",
		"Mon, 2 Jan 2006 15:04:05 -0700 (MST)",
		"2 Jan 2006 15:04:05 -0700",
		"Mon, 02 Jan 2006 15:04:05 -0700",
		time.RFC3339,
	}
	s = strings.TrimSpace(s)
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t
		}
	}
	return time.Now()
}

func decodeHeader(s string) string {
	return s
}

func ProviderDefaults(provider string) (imapHost string, imapPort int, smtpHost string, smtpPort int) {
	switch strings.ToLower(provider) {
	case "gmail":
		return "imap.gmail.com", 993, "smtp.gmail.com", 587
	case "outlook", "hotmail":
		return "outlook.office365.com", 993, "smtp-mail.outlook.com", 587
	case "yahoo":
		return "imap.mail.yahoo.com", 993, "smtp.mail.yahoo.com", 587
	case "mailpit":
		return "mailpit", 1143, "mailpit", 1025
	case "fake-imap":
		return "fake-imap", 1993, "mailpit", 1025
	default:
		return "", 993, "", 587
	}
}
