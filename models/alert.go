package models

import "time"

type GridAlert struct {
	Type      string    `json:"type"`      // ALWAYS "ALERT"
	Severity  string    `json:"severity"`  // "INFO", "WARNING", "CRITICAL"
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

type TelemetryMessage struct {
	Type      string       `json:"type"` // "TELEMETRY"
	Summary   *GridSummary `json:"summary"`
	Timestamp time.Time    `json:"timestamp"`
}
