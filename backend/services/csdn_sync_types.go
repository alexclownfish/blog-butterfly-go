package services

import "time"

type CSDNSyncSessionStatus string

const (
	CSDNSyncSessionStatusPending    CSDNSyncSessionStatus = "pending"
	CSDNSyncSessionStatusScanned    CSDNSyncSessionStatus = "scanned"
	CSDNSyncSessionStatusAuthorized CSDNSyncSessionStatus = "authorized"
	CSDNSyncSessionStatusExpired    CSDNSyncSessionStatus = "expired"
	CSDNSyncSessionStatusFailed     CSDNSyncSessionStatus = "failed"
)

type CSDNSyncSession struct {
	ID              string                 `json:"id"`
	UserID          uint                   `json:"user_id,omitempty"`
	Provider        string                 `json:"provider"`
	ProviderMode    string                 `json:"provider_mode"`
	ProviderSession string                 `json:"provider_session,omitempty"`
	Status          CSDNSyncSessionStatus  `json:"status"`
	Message         string                 `json:"message,omitempty"`
	ErrorMessage    string                 `json:"error_message,omitempty"`
	QRCodeDataURL   string                 `json:"qr_code_data_url,omitempty"`
	ExpiresAt       time.Time              `json:"expires_at"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	Articles        []CSDNSyncRemoteArticle `json:"articles,omitempty"`
}

type CSDNSyncRemoteArticle struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Summary     string    `json:"summary,omitempty"`
	CoverImage  string    `json:"cover_image,omitempty"`
	SourceURL   string    `json:"source_url,omitempty"`
	PublishedAt time.Time `json:"published_at,omitempty"`
}

type CSDNSyncLoginStartResult struct {
	Provider        string
	ProviderMode    string
	ProviderSession string
	QRCodeDataURL   string
	Message         string
}
