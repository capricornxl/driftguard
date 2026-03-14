package alerter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/driftguard/driftguard/pkg/config"
	"github.com/driftguard/driftguard/pkg/models"
	"github.com/sirupsen/logrus"
)

// Alerter sends alerts through various channels
type Alerter struct {
	config  *config.AlerterConfig
	logger  *logrus.Logger
	channels []AlertChannel
}

// AlertChannel interface for different alert destinations
type AlertChannel interface {
	Send(alert *models.Alert) error
	Name() string
}

// NewAlerter creates a new alerter instance
func NewAlerter(cfg *config.AlerterConfig, logger *logrus.Logger) *Alerter {
	a := &Alerter{
		config:   cfg,
		logger:   logger,
		channels: []AlertChannel{},
	}

	// Initialize configured channels
	for _, channelCfg := range cfg.Channels {
		if !channelCfg.Enabled {
			continue
		}

		switch channelCfg.Type {
		case "log":
			a.channels = append(a.channels, NewLogChannel(logger))
		case "slack":
			if channelCfg.Webhook != "" {
				a.channels = append(a.channels, NewSlackChannel(channelCfg.Webhook, logger))
			}
		case "discord":
			if channelCfg.Webhook != "" {
				a.channels = append(a.channels, NewDiscordChannel(channelCfg.Webhook, logger))
			}
		case "webhook":
			if channelCfg.URL != "" {
				a.channels = append(a.channels, NewWebhookChannel(channelCfg.URL, logger))
			}
		}
	}

	return a
}

// SendAlert sends an alert through all configured channels with retry mechanism
func (a *Alerter) SendAlert(alert *models.Alert) error {
	a.logger.WithFields(logrus.Fields{
		"alert_id":   alert.ID,
		"agent_id":   alert.AgentID,
		"type":       alert.Type,
		"severity":   alert.Severity,
		"message":    alert.Message,
	}).Info("Sending alert")

	errors := []error{}
	for _, channel := range a.channels {
		if err := a.sendWithRetry(channel, alert); err != nil {
			a.logger.WithFields(logrus.Fields{
				"channel": channel.Name(),
				"error":   err,
			}).Error("Failed to send alert after retries")
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to send alerts to %d/%d channels", len(errors), len(a.channels))
	}

	return nil
}

// sendWithRetry sends an alert with exponential backoff retry
func (a *Alerter) sendWithRetry(channel AlertChannel, alert *models.Alert) error {
	maxRetries := 3
	baseDelay := time.Second

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		err := channel.Send(alert)
		if err == nil {
			if attempt > 0 {
				a.logger.WithFields(logrus.Fields{
					"channel": channel.Name(),
					"attempt": attempt + 1,
				}).Info("Alert sent successfully after retry")
			}
			return nil
		}

		lastErr = err

		if attempt < maxRetries-1 {
			// Exponential backoff: 1s, 2s, 4s
			delay := baseDelay * time.Duration(1<<uint(attempt))
			a.logger.WithFields(logrus.Fields{
				"channel": channel.Name(),
				"attempt": attempt + 1,
				"delay":   delay,
				"error":   err,
			}).Warn("Alert send failed, retrying with exponential backoff")
			time.Sleep(delay)
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

// ============ Log Channel ============

type logChannel struct {
	logger *logrus.Logger
}

func NewLogChannel(logger *logrus.Logger) AlertChannel {
	return &logChannel{logger: logger}
}

func (c *logChannel) Name() string {
	return "log"
}

func (c *logChannel) Send(alert *models.Alert) error {
	c.logger.WithFields(logrus.Fields{
		"alert_id":  alert.ID,
		"agent_id":  alert.AgentID,
		"type":      alert.Type,
		"severity":  alert.Severity,
		"message":   alert.Message,
		"created_at": alert.CreatedAt,
	}).Info("ALERT")
	return nil
}

// ============ Slack Channel ============

type slackChannel struct {
	webhookURL string
	logger     *logrus.Logger
}

func NewSlackChannel(webhookURL string, logger *logrus.Logger) AlertChannel {
	return &slackChannel{
		webhookURL: webhookURL,
		logger:     logger,
	}
}

func (c *slackChannel) Name() string {
	return "slack"
}

func (c *slackChannel) Send(alert *models.Alert) error {
	// Determine color based on severity
	color := "#36a64f" // green
	emoji := "✅"
	switch alert.Severity {
	case "info":
		color = "#36a64f"
		emoji = "ℹ️"
	case "warning":
		color = "#ff9800"
		emoji = "⚠️"
	case "critical":
		color = "#ff0000"
		emoji = "🚨"
	}

	// Create Slack attachment
	attachment := SlackAttachment{
		Color:      color,
		Title:      fmt.Sprintf("%s DriftGuard Alert: %s", emoji, alert.Type),
		TitleLink:  "",
		Fallback:   alert.Message,
		Text:       alert.Message,
		Fields: []SlackField{
			{
				Title: "Agent",
				Value: alert.AgentID,
				Short: true,
			},
			{
				Title: "Severity",
				Value: alert.Severity,
				Short: true,
			},
			{
				Title: "Status",
				Value: alert.Status,
				Short: true,
			},
			{
				Title: "Time",
				Value: alert.CreatedAt.Format(time.RFC3339),
				Short: true,
			},
		},
		Footer: "DriftGuard",
		Ts:     json.Number(fmt.Sprintf("%d", alert.CreatedAt.Unix())),
	}

	payload := SlackPayload{
		Attachments: []SlackAttachment{attachment},
		Username:    "DriftGuard",
		IconEmoji:   ":robot_face:",
	}

	return c.sendWebhook(payload)
}

type SlackPayload struct {
	Text        string            `json:"text,omitempty"`
	Username    string            `json:"username,omitempty"`
	IconEmoji   string            `json:"icon_emoji,omitempty"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
}

type SlackAttachment struct {
	Color      string       `json:"color,omitempty"`
	Title      string       `json:"title,omitempty"`
	TitleLink  string       `json:"title_link,omitempty"`
	Text       string       `json:"text,omitempty"`
	Fallback   string       `json:"fallback,omitempty"`
	Fields     []SlackField `json:"fields,omitempty"`
	Footer     string       `json:"footer,omitempty"`
	Ts         json.Number  `json:"ts,omitempty"`
}

type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// ============ Discord Channel ============

type discordChannel struct {
	webhookURL string
	logger     *logrus.Logger
}

func NewDiscordChannel(webhookURL string, logger *logrus.Logger) AlertChannel {
	return &discordChannel{
		webhookURL: webhookURL,
		logger:     logger,
	}
}

func (c *discordChannel) Name() string {
	return "discord"
}

func (c *discordChannel) Send(alert *models.Alert) error {
	// Determine color based on severity
	color := 3066993 // green
	emoji := "✅"
	switch alert.Severity {
	case "info":
		color = 3066993
		emoji = "ℹ️"
	case "warning":
		color = 15105570
		emoji = "⚠️"
	case "critical":
		color = 15158332
		emoji = "🚨"
	}

	// Create Discord embed
	embed := DiscordEmbed{
		Title:       fmt.Sprintf("%s DriftGuard Alert: %s", emoji, alert.Type),
		Description: alert.Message,
		Color:       color,
		Fields: []DiscordEmbedField{
			{
				Name:   "Agent",
				Value:  alert.AgentID,
				Inline: true,
			},
			{
				Name:   "Severity",
				Value:  alert.Severity,
				Inline: true,
			},
			{
				Name:   "Status",
				Value:  alert.Status,
				Inline: true,
			},
			{
				Name:   "Time",
				Value:  alert.CreatedAt.Format(time.RFC3339),
				Inline: false,
			},
		},
		Footer: &DiscordEmbedFooter{
			Text: "DriftGuard",
		},
		Timestamp: alert.CreatedAt.Format(time.RFC3339),
	}

	payload := DiscordPayload{
		Username:  "DriftGuard",
		AvatarURL: "",
		Embeds:    []DiscordEmbed{embed},
	}

	return c.sendWebhook(payload)
}

type DiscordPayload struct {
	Username  string         `json:"username,omitempty"`
	AvatarURL string         `json:"avatar_url,omitempty"`
	Embeds    []DiscordEmbed `json:"embeds,omitempty"`
}

type DiscordEmbed struct {
	Title       string               `json:"title,omitempty"`
	Description string               `json:"description,omitempty"`
	Color       int                  `json:"color,omitempty"`
	Fields      []DiscordEmbedField  `json:"fields,omitempty"`
	Footer      *DiscordEmbedFooter  `json:"footer,omitempty"`
	Timestamp   string               `json:"timestamp,omitempty"`
}

type DiscordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type DiscordEmbedFooter struct {
	Text string `json:"text"`
}

// ============ Webhook Channel ============

type webhookChannel struct {
	url    string
	logger *logrus.Logger
}

func NewWebhookChannel(url string, logger *logrus.Logger) AlertChannel {
	return &webhookChannel{
		url:    url,
		logger: logger,
	}
}

func (c *webhookChannel) Name() string {
	return "webhook"
}

func (c *webhookChannel) Send(alert *models.Alert) error {
	// Generic JSON payload
	payload := map[string]interface{}{
		"alert_id":   alert.ID,
		"agent_id":   alert.AgentID,
		"type":       alert.Type,
		"severity":   alert.Severity,
		"status":     alert.Status,
		"message":    alert.Message,
		"created_at": alert.CreatedAt.Format(time.RFC3339),
	}

	return c.sendJSON(payload)
}

// ============ Common Methods ============

func (c *slackChannel) sendWebhook(payload SlackPayload) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := http.Post(c.webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	c.logger.Debug("Slack alert sent successfully")
	return nil
}

func (c *discordChannel) sendWebhook(payload DiscordPayload) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := http.Post(c.webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 && resp.StatusCode != 200 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	c.logger.Debug("Discord alert sent successfully")
	return nil
}

func (c *webhookChannel) sendJSON(payload map[string]interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := http.Post(c.url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	c.logger.Debug("Webhook alert sent successfully")
	return nil
}
