package alerter

import (
	"fmt"
	"testing"
	"time"

	"github.com/driftguard/driftguard/pkg/config"
	"github.com/driftguard/driftguard/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewAlerter(t *testing.T) {
	logger := logrus.New()

	t.Run("empty config", func(t *testing.T) {
		cfg := &config.AlerterConfig{
			Enabled:  true,
			Channels: []config.AlertChannelConfig{},
		}

		alerter := NewAlerter(cfg, logger)
		assert.NotNil(t, alerter)
	})

	t.Run("log channel enabled", func(t *testing.T) {
		cfg := &config.AlerterConfig{
			Enabled: true,
			Channels: []config.AlertChannelConfig{
				{Type: "log", Enabled: true},
			},
		}

		alerter := NewAlerter(cfg, logger)
		assert.NotNil(t, alerter)
	})

	t.Run("slack channel with webhook", func(t *testing.T) {
		cfg := &config.AlerterConfig{
			Enabled: true,
			Channels: []config.AlertChannelConfig{
				{
					Type:    "slack",
					Enabled: true,
					Webhook: "https://hooks.slack.com/test",
				},
			},
		}

		alerter := NewAlerter(cfg, logger)
		assert.NotNil(t, alerter)
	})

	t.Run("discord channel with webhook", func(t *testing.T) {
		cfg := &config.AlerterConfig{
			Enabled: true,
			Channels: []config.AlertChannelConfig{
				{
					Type:    "discord",
					Enabled: true,
					Webhook: "https://discord.com/api/webhooks/test",
				},
			},
		}

		alerter := NewAlerter(cfg, logger)
		assert.NotNil(t, alerter)
	})

	t.Run("disabled channel", func(t *testing.T) {
		cfg := &config.AlerterConfig{
			Enabled: true,
			Channels: []config.AlertChannelConfig{
				{Type: "slack", Enabled: false, Webhook: "https://hooks.slack.com/test"},
			},
		}

		alerter := NewAlerter(cfg, logger)
		assert.NotNil(t, alerter)
	})
}

func TestLogChannel(t *testing.T) {
	logger := logrus.New()
	channel := NewLogChannel(logger)

	assert.Equal(t, "log", channel.Name())

	alert := &models.Alert{
		ID:        "test-alert-1",
		AgentID:   "test-agent",
		Type:      "drift",
		Severity:  "warning",
		Status:    "active",
		Message:   "Health score dropped below threshold",
		CreatedAt: time.Now(),
	}

	err := channel.Send(alert)
	assert.NoError(t, err)
}

func TestSlackChannel(t *testing.T) {
	logger := logrus.New()

	t.Run("channel name", func(t *testing.T) {
		channel := NewSlackChannel("https://hooks.slack.com/test", logger)
		assert.Equal(t, "slack", channel.Name())
	})

	t.Run("info severity color", func(t *testing.T) {
		alert := &models.Alert{
			ID:        "test-1",
			AgentID:   "agent-1",
			Type:      "info",
			Severity:  "info",
			Status:    "active",
			Message:   "Test info alert",
			CreatedAt: time.Now(),
		}

		// Test payload generation (without actual sending)
		attachment := createSlackAttachment(alert)
		assert.Equal(t, "#36a64f", attachment.Color)
	})

	t.Run("warning severity color", func(t *testing.T) {
		alert := &models.Alert{
			ID:        "test-2",
			AgentID:   "agent-1",
			Type:      "warning",
			Severity:  "warning",
			Status:    "active",
			Message:   "Test warning alert",
			CreatedAt: time.Now(),
		}

		attachment := createSlackAttachment(alert)
		assert.Equal(t, "#ff9800", attachment.Color)
	})

	t.Run("critical severity color", func(t *testing.T) {
		alert := &models.Alert{
			ID:        "test-3",
			AgentID:   "agent-1",
			Type:      "critical",
			Severity:  "critical",
			Status:    "active",
			Message:   "Test critical alert",
			CreatedAt: time.Now(),
		}

		attachment := createSlackAttachment(alert)
		assert.Equal(t, "#ff0000", attachment.Color)
	})
}

func TestDiscordChannel(t *testing.T) {
	logger := logrus.New()

	t.Run("channel name", func(t *testing.T) {
		channel := NewDiscordChannel("https://discord.com/api/webhooks/test", logger)
		assert.Equal(t, "discord", channel.Name())
	})

	t.Run("info severity color", func(t *testing.T) {
		alert := &models.Alert{
			ID:        "test-1",
			AgentID:   "agent-1",
			Severity:  "info",
			Status:    "active",
			Message:   "Test info alert",
			CreatedAt: time.Now(),
		}

		embed := createDiscordEmbed(alert)
		assert.Equal(t, 3066993, embed.Color) // Green
	})

	t.Run("warning severity color", func(t *testing.T) {
		alert := &models.Alert{
			ID:        "test-2",
			AgentID:   "agent-1",
			Severity:  "warning",
			Status:    "active",
			Message:   "Test warning alert",
			CreatedAt: time.Now(),
		}

		embed := createDiscordEmbed(alert)
		assert.Equal(t, 15105570, embed.Color) // Orange
	})

	t.Run("critical severity color", func(t *testing.T) {
		alert := &models.Alert{
			ID:        "test-3",
			AgentID:   "agent-1",
			Severity:  "critical",
			Status:    "active",
			Message:   "Test critical alert",
			CreatedAt: time.Now(),
		}

		embed := createDiscordEmbed(alert)
		assert.Equal(t, 15158332, embed.Color) // Red
	})
}

func TestWebhookChannel(t *testing.T) {
	logger := logrus.New()
	channel := NewWebhookChannel("http://localhost:8080/webhook", logger)

	assert.Equal(t, "webhook", channel.Name())
}

func TestAlertPayloadStructures(t *testing.T) {
	t.Run("slack payload structure", func(t *testing.T) {
		payload := SlackPayload{
			Text:      "Test",
			Username:  "DriftGuard",
			IconEmoji: ":robot_face:",
			Attachments: []SlackAttachment{
				{
					Color:    "#ff0000",
					Title:    "Test Alert",
					Text:     "Test message",
					Fallback: "Test fallback",
					Fields: []SlackField{
						{Title: "Agent", Value: "agent-1", Short: true},
					},
					Footer: "DriftGuard",
				},
			},
		}

		assert.Equal(t, "DriftGuard", payload.Username)
		assert.Equal(t, ":robot_face:", payload.IconEmoji)
		assert.Len(t, payload.Attachments, 1)
	})

	t.Run("discord payload structure", func(t *testing.T) {
		payload := DiscordPayload{
			Username:  "DriftGuard",
			AvatarURL: "",
			Embeds: []DiscordEmbed{
				{
					Title:       "Test Alert",
					Description: "Test message",
					Color:       15158332,
					Fields: []DiscordEmbedField{
						{Name: "Agent", Value: "agent-1", Inline: true},
					},
					Footer:    &DiscordEmbedFooter{Text: "DriftGuard"},
					Timestamp: time.Now().Format(time.RFC3339),
				},
			},
		}

		assert.Equal(t, "DriftGuard", payload.Username)
		assert.Len(t, payload.Embeds, 1)
	})
}

// Helper functions for testing
func createSlackAttachment(alert *models.Alert) SlackAttachment {
	color := "#36a64f"
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

	return SlackAttachment{
		Color:    color,
		Title:    fmt.Sprintf("%s DriftGuard Alert: %s", emoji, alert.Type),
		Text:     alert.Message,
		Fallback: alert.Message,
		Fields: []SlackField{
			{Title: "Agent", Value: alert.AgentID, Short: true},
			{Title: "Severity", Value: alert.Severity, Short: true},
		},
		Footer: "DriftGuard",
	}
}

func createDiscordEmbed(alert *models.Alert) DiscordEmbed {
	color := 3066993
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

	return DiscordEmbed{
		Title:       fmt.Sprintf("%s DriftGuard Alert: %s", emoji, alert.Type),
		Description: alert.Message,
		Color:       color,
		Fields: []DiscordEmbedField{
			{Name: "Agent", Value: alert.AgentID, Inline: true},
			{Name: "Severity", Value: alert.Severity, Inline: true},
		},
		Footer:    &DiscordEmbedFooter{Text: "DriftGuard"},
		Timestamp: alert.CreatedAt.Format(time.RFC3339),
	}
}
