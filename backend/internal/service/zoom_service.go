package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ZoomTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type ZoomMeetingRequest struct {
	Topic     string              `json:"topic"`
	Type      int                 `json:"type"`
	StartTime string              `json:"start_time"`
	Duration  int                 `json:"duration"`
	Timezone  string              `json:"timezone"`
	Agenda    string              `json:"agenda,omitempty"`
	Settings  ZoomMeetingSettings `json:"settings"`
}

type ZoomMeetingSettings struct {
	HostVideo        bool   `json:"host_video"`
	ParticipantVideo bool   `json:"participant_video"`
	JoinBeforeHost   bool   `json:"join_before_host"`
	MuteUponEntry    bool   `json:"mute_upon_entry"`
	AutoRecording    string `json:"auto_recording"`
	WaitingRoom      bool   `json:"waiting_room"`
}

type ZoomMeetingResponse struct {
	ID        int64  `json:"id"`
	Topic     string `json:"topic"`
	StartTime string `json:"start_time"`
	Duration  int    `json:"duration"`
	JoinURL   string `json:"join_url"`
	StartURL  string `json:"start_url"`
	Password  string `json:"password"`
	Status    string `json:"status"`
}

type ZoomUserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Type  int    `json:"type"`
}

type ZoomErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func GetZoomAccessToken(accountID, clientID, clientSecret string) (string, error) {
	data := url.Values{}
	data.Set("grant_type", "account_credentials")
	data.Set("account_id", accountID)

	req, err := http.NewRequest("POST", "https://zoom.us/oauth/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	credentials := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	req.Header.Set("Authorization", "Basic "+credentials)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("zoom token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var zoomErr ZoomErrorResponse
		json.Unmarshal(body, &zoomErr)
		return "", fmt.Errorf("zoom auth failed (HTTP %d): %s", resp.StatusCode, zoomErr.Message)
	}

	var tokenResp ZoomTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse zoom token: %w", err)
	}

	return tokenResp.AccessToken, nil
}

func CreateZoomMeeting(accessToken, topic, agenda string, startTime time.Time, durationMin int) (*ZoomMeetingResponse, error) {
	meeting := ZoomMeetingRequest{
		Topic:     topic,
		Type:      2,
		StartTime: startTime.UTC().Format("2006-01-02T15:04:05Z"),
		Duration:  durationMin,
		Timezone:  "UTC",
		Agenda:    agenda,
		Settings: ZoomMeetingSettings{
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
		var zoomErr ZoomErrorResponse
		json.Unmarshal(body, &zoomErr)
		slog.Error("Zoom create meeting error", "status", resp.StatusCode, "body", string(body))
		return nil, fmt.Errorf("zoom API error (HTTP %d): %s", resp.StatusCode, zoomErr.Message)
	}

	var meetingResp ZoomMeetingResponse
	if err := json.Unmarshal(body, &meetingResp); err != nil {
		return nil, fmt.Errorf("failed to parse zoom response: %w", err)
	}

	return &meetingResp, nil
}

func UpdateZoomMeeting(accessToken string, meetingID int64, topic, agenda string, startTime time.Time, durationMin int) error {
	meeting := ZoomMeetingRequest{
		Topic:     topic,
		Type:      2,
		StartTime: startTime.UTC().Format("2006-01-02T15:04:05Z"),
		Duration:  durationMin,
		Timezone:  "UTC",
		Agenda:    agenda,
		Settings: ZoomMeetingSettings{
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

func DeleteZoomMeeting(accessToken string, meetingID int64) error {
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

func GetZoomUser(accessToken string) (*ZoomUserResponse, error) {
	req, err := http.NewRequest("GET", "https://api.zoom.us/v2/users/me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("zoom get user failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("zoom user error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var user ZoomUserResponse
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, err
	}

	return &user, nil
}
