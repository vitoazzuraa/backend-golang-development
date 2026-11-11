package model

import "encoding/json"

const (
	EventEnrollmentCreated = "enrollment.created"
	EventEnrollmentDropped = "enrollment.dropped"
	EventEnrollmentSwapped = "enrollment.swapped"
	EventQuotaChanged      = "course.quota_changed"
)

const (
	TaskSendNotification = "notification:send"
	QueueNotifications   = "notifications"
)

type EnrollmentEvent struct {
	UserID         int    `json:"user_id,omitempty"`
	CourseID       int    `json:"course_id"`
	CourseCode     string `json:"course_code"`
	CourseName     string `json:"course_name"`
	FromCourseID   int    `json:"from_course_id,omitempty"`
	FromCourseCode string `json:"from_course_code,omitempty"`
	FromCourseName string `json:"from_course_name,omitempty"`
	Quota          int    `json:"quota,omitempty"`
}

type NotificationJob struct {
	OutboxID  int64           `json:"outbox_id"`
	EventType string          `json:"event_type"`
	Payload   json.RawMessage `json:"payload"`
}

type Message struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type QueueInfo struct {
	Name      string `json:"name"`
	Pending   int    `json:"pending"`
	Active    int    `json:"active"`
	Scheduled int    `json:"scheduled"`
	Retry     int    `json:"retry"`
	Archived  int    `json:"archived"`
	Completed int    `json:"completed"`
	Processed int    `json:"processed_today"`
	Failed    int    `json:"failed_today"`
}

type QueueStats struct {
	OutboxPending        int        `json:"outbox_pending"`
	OutboxTotal          int        `json:"outbox_total"`
	NotifikasiBelumKirim int        `json:"notifikasi_belum_kirim"`
	NotifikasiTotal      int        `json:"notifikasi_total"`
	Queue                *QueueInfo `json:"queue,omitempty"`
	QueueError           string     `json:"queue_error"`
}
