package model

import "time"

type AppSettings struct {
	AppDomain           string    `json:"appDomain"`
	RegistrationEnabled bool      `json:"registrationEnabled"`
	SMTPHost            string    `json:"smtpHost"`
	SMTPPort            string    `json:"smtpPort"`
	SMTPUsername        string    `json:"smtpUsername"`
	SMTPPassword        string    `json:"-"`
	SMTPPasswordSet     bool      `json:"smtpPasswordSet"`
	SMTPFrom            string    `json:"smtpFrom"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}
