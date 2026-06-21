package zoom

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type MeetingRequest struct {
	Topic     string          `json:"topic"`
	Type      int             `json:"type"`
	StartTime string          `json:"start_time"`
	Duration  int             `json:"duration"`
	Timezone  string          `json:"timezone"`
	Agenda    string          `json:"agenda,omitempty"`
	Settings  MeetingSettings `json:"settings"`
}

type MeetingSettings struct {
	HostVideo        bool   `json:"host_video"`
	ParticipantVideo bool   `json:"participant_video"`
	JoinBeforeHost   bool   `json:"join_before_host"`
	MuteUponEntry    bool   `json:"mute_upon_entry"`
	AutoRecording    string `json:"auto_recording"`
	WaitingRoom      bool   `json:"waiting_room"`
}

type MeetingResponse struct {
	ID        int64  `json:"id"`
	Topic     string `json:"topic"`
	StartTime string `json:"start_time"`
	Duration  int    `json:"duration"`
	JoinURL   string `json:"join_url"`
	StartURL  string `json:"start_url"`
	Password  string `json:"password"`
	Status    string `json:"status"`
}

type UserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Type  int    `json:"type"`
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func debugLog(method string, kvPairs ...interface{}) {
	args := make([]interface{}, 0, len(kvPairs)+1)
	args = append(args, method)
	args = append(args, kvPairs...)
	log.Println(args...)
}

func debugLogResult(method string, err error, kvPairs ...interface{}) {
	if err != nil {
		args := make([]interface{}, 0, len(kvPairs)+2)
		args = append(args, method, "ERROR:", err)
		args = append(args, kvPairs...)
		log.Println(args...)
		return
	}
	args := make([]interface{}, 0, len(kvPairs)+2)
	args = append(args, method, "OK")
	args = append(args, kvPairs...)
	log.Println(args...)
}

func GetAccessToken(accountID, clientID, clientSecret string) (string, error) {
	debugLog("Zoom.GetAccessToken", "account_id", accountID, "client_id_len", len(clientID), "secret_len", len(clientSecret))

	data := url.Values{}
	data.Set("grant_type", "account_credentials")
	data.Set("account_id", accountID)

	req, err := http.NewRequest("POST", "https://zoom.us/oauth/token", strings.NewReader(data.Encode()))
	if err != nil {
		debugLogResult("Zoom.GetAccessToken", err, "step", "build_request")
		return "", err
	}

	credentials := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	req.Header.Set("Authorization", "Basic "+credentials)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 15 * time.Second}
	debugLog("Zoom.GetAccessToken", "step", "sending_token_request")
	resp, err := client.Do(req)
	if err != nil {
		debugLogResult("Zoom.GetAccessToken", err, "step", "http_request")
		return "", fmt.Errorf("zoom token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	debugLog("Zoom.GetAccessToken", "step", "response_received", "status", resp.StatusCode, "body_len", len(body))

	if resp.StatusCode != http.StatusOK {
		var zoomErr ErrorResponse
		json.Unmarshal(body, &zoomErr)
		debugLogResult("Zoom.GetAccessToken", fmt.Errorf("HTTP %d: %s", resp.StatusCode, zoomErr.Message), "step", "auth_failed", "response_body", string(body))
		return "", fmt.Errorf("zoom auth failed (HTTP %d): %s", resp.StatusCode, zoomErr.Message)
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		debugLogResult("Zoom.GetAccessToken", err, "step", "parse_token")
		return "", fmt.Errorf("failed to parse zoom token: %w", err)
	}

	debugLogResult("Zoom.GetAccessToken", nil, "token_len", len(tokenResp.AccessToken), "expires_in", tokenResp.ExpiresIn)
	return tokenResp.AccessToken, nil
}

func CreateMeeting(accessToken, topic, agenda string, startTime time.Time, durationMin int) (*MeetingResponse, error) {
	meeting := MeetingRequest{
		Topic:     topic,
		Type:      2,
		StartTime: startTime.UTC().Format("2006-01-02T15:04:05Z"),
		Duration:  durationMin,
		Timezone:  "UTC",
		Agenda:    agenda,
		Settings: MeetingSettings{
			HostVideo:        true,
			ParticipantVideo: true,
			JoinBeforeHost:   true,
			MuteUponEntry:    false,
			AutoRecording:    "none",
			WaitingRoom:      false,
		},
	}

	payload, err := json.Marshal(meeting)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.zoom.us/v2/users/me/meetings", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("zoom create meeting failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		var zoomErr ErrorResponse
		json.Unmarshal(body, &zoomErr)
		slog.Error("Zoom create meeting error", "status", resp.StatusCode, "body", string(body))
		return nil, fmt.Errorf("zoom API error (HTTP %d): %s", resp.StatusCode, zoomErr.Message)
	}

	var meetingResp MeetingResponse
	if err := json.Unmarshal(body, &meetingResp); err != nil {
		return nil, fmt.Errorf("failed to parse zoom response: %w", err)
	}

	return &meetingResp, nil
}

func UpdateMeeting(accessToken string, meetingID int64, topic, agenda string, startTime time.Time, durationMin int) error {
	meeting := MeetingRequest{
		Topic:     topic,
		Type:      2,
		StartTime: startTime.UTC().Format("2006-01-02T15:04:05Z"),
		Duration:  durationMin,
		Timezone:  "UTC",
		Agenda:    agenda,
		Settings: MeetingSettings{
			HostVideo:        true,
			ParticipantVideo: true,
			JoinBeforeHost:   true,
			MuteUponEntry:    false,
			AutoRecording:    "none",
			WaitingRoom:      false,
		},
	}

	payload, _ := json.Marshal(meeting)

	req, err := http.NewRequest("PATCH", fmt.Sprintf("https://api.zoom.us/v2/meetings/%d", meetingID), bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("zoom update meeting failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("zoom update error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

func DeleteMeeting(accessToken string, meetingID int64) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("https://api.zoom.us/v2/meetings/%d", meetingID), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("zoom delete meeting failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("zoom delete error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

func GetUser(accessToken string) (*UserResponse, error) {
	debugLog("Zoom.GetUser", "step", "start", "token_len", len(accessToken))

	req, err := http.NewRequest("GET", "https://api.zoom.us/v2/users/me", nil)
	if err != nil {
		debugLogResult("Zoom.GetUser", err, "step", "build_request")
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		debugLogResult("Zoom.GetUser", err, "step", "http_request")
		return nil, fmt.Errorf("zoom get user failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	debugLog("Zoom.GetUser", "step", "response_received", "status", resp.StatusCode, "body_len", len(body))

	if resp.StatusCode != http.StatusOK {
		debugLogResult("Zoom.GetUser", fmt.Errorf("HTTP %d", resp.StatusCode), "step", "api_error", "response_body", string(body))
		return nil, fmt.Errorf("zoom user error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var user UserResponse
	if err := json.Unmarshal(body, &user); err != nil {
		debugLogResult("Zoom.GetUser", err, "step", "parse_response")
		return nil, err
	}

	debugLogResult("Zoom.GetUser", nil, "email", user.Email, "zoom_user_id", user.ID)
	return &user, nil
}
