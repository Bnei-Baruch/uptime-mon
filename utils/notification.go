package utils

import (
	"bytes"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
	"net/smtp"
	"strconv"
	"time"
)

type SlackMessage struct {
	Username  string `json:"username,omitempty"`
	IconEmoji string `json:"icon_emoji,omitempty"`
	Channel   string `json:"channel,omitempty"`
	Text      string `json:"text,omitempty"`
}

func SendEmail(subject string, body string) {
	log.Debugf("MQTT: Sending mail..\n")

	user := viper.GetString("notification.mail_user")
	password := viper.GetString("notification.mail_pass")
	smtpHost := viper.GetString("notification.mail_host")
	from := viper.GetString("notification.mail_from")
	to := viper.GetStringSlice("notification.mail_to")
	smtpPort := "587"

	msg := []byte("From: " + from + "\r\n" + "Subject: " + subject + "\r\n" + "\r\n" + "Product ID: " + body + "\r\n")
	auth := smtp.PlainAuth("", user, password, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, msg)
	if err != nil {
		log.Errorf("MAIL: Send mail error: %s", err)
	} else {
		log.Debugf("MAIL: Mail sent. \n")
	}
}

func SendSlack(msg string) {
	webhookUrl := viper.GetString("notification.slack_webhook")

	slackBody, _ := json.Marshal(SlackMessage{Text: msg})
	req, err := http.NewRequest(http.MethodPost, webhookUrl, bytes.NewBuffer(slackBody))
	if err != nil {
		return
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		log.Errorf("SLACK: Send slack error: %s", err)
		return
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Errorf("SLACK: Send slack bad status code: %s", strconv.Itoa(res.StatusCode))
		return
	}
}
