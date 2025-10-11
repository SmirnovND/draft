# Services

Слой вспомогательных сервисов (email, SMS, внешние API и т.д.).

## Пример сервиса

```go
package services

import (
	"context"
	"fmt"
	"net/smtp"
)

type EmailService struct {
	smtpHost string
	smtpPort string
	from     string
	password string
}

func NewEmailService(host, port, from, password string) *EmailService {
	return &EmailService{
		smtpHost: host,
		smtpPort: port,
		from:     from,
		password: password,
	}
}

func (s *EmailService) SendEmail(ctx context.Context, to, subject, body string) error {
	auth := smtp.PlainAuth("", s.from, s.password, s.smtpHost)
	
	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", to, subject, body))
	
	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)
	return smtp.SendMail(addr, auth, s.from, []string{to}, msg)
}
```

## Регистрация в DI контейнере

В файле `internal/container/container.go`:

```go
func (c *Container) provideService() {
	c.container.Provide(func() *services.EmailService {
		return services.NewEmailService(
			"smtp.gmail.com",
			"587",
			"your-email@gmail.com",
			"your-password",
		)
	})
}
```