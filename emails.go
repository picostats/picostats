package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	ttemplate "text/template"
	"time"

	"github.com/SparkPost/gosparkpost"
	"github.com/abiosoft/semaphore"
)

type EmailManager struct {
	EmailDataObjects []*EmailData
	Client           *gosparkpost.Client
	Semaphore        *semaphore.Semaphore
}

func initEmails() {
	em = &EmailManager{}

	cfg := &gosparkpost.Config{
		BaseUrl:    "https://api.sparkpost.com",
		ApiKey:     conf.SparkpostKey,
		ApiVersion: 1,
	}

	em.Client = &gosparkpost.Client{}

	err := em.Client.Init(cfg)
	if err != nil {
		log.Printf("[emails.go] Init error: %s", err)
	}

	em.Semaphore = semaphore.New(32)

	em.StartSending()
}

func (e *EmailManager) Queue(ed *EmailData) {
	e.EmailDataObjects = append(e.EmailDataObjects, ed)
}

func (e *EmailManager) StartSending() {
	go func() {
		for {
			for len(e.EmailDataObjects) > 0 {
				e.Semaphore.Acquire()
				var ed *EmailData
				ed, e.EmailDataObjects = e.EmailDataObjects[0], e.EmailDataObjects[1:]
				go e.Send(ed)
			}
			time.Sleep(time.Second * 3)
		}
	}()
}

func (e *EmailManager) Send(ed *EmailData) error {
	content := gosparkpost.Content{
		HTML:    ed.EmailMessage.HTML,
		From:    fmt.Sprintf("%s <%s>", ed.EmailMessage.FromName, ed.EmailMessage.FromEmail),
		Subject: ed.EmailMessage.Subject,
		Text:    ed.EmailMessage.Text,
	}

	if len(ed.EmailMessage.ReplyTo) > 0 {
		content.ReplyTo = ed.EmailMessage.ReplyTo
	}

	tx := &gosparkpost.Transmission{
		Recipients: []string{ed.EmailMessage.To},
		Content:    content,
	}

	_, _, err := e.Client.Send(tx)
	if err == nil {
		log.Printf("[emails.go] Email sent: %s %s %s", ed.EmailMessage.To, ed.EmailMessage.FromEmail, ed.EmailMessage.Subject)
	} else {
		log.Printf("[emails.go] err: %s %s %s %s", err, ed.EmailMessage.To, ed.EmailMessage.FromEmail, ed.EmailMessage.Subject)
	}

	e.Semaphore.Release()

	return err
}

type EmailMessage struct {
	Subject   string
	Text      string
	HTML      string
	ReplyTo   string
	FromName  string
	FromEmail string
	To        string
}

type EmailData struct {
	VerificationLink string
	EmailMessage     *EmailMessage
}

func ParseHTML(ed *EmailData) string {
	buffer := new(bytes.Buffer)

	t, err := template.New("welcome.html").ParseFiles("templates/emails/welcome.html")
	if err != nil {
		log.Printf("[emails.go] ParseFIles err: %s", err)
	}

	err = t.Execute(buffer, ed)
	if err != nil {
		log.Printf("[emails.go] Execute err: %s", err)
	}

	return buffer.String()
}

func ParseText(ed *EmailData) string {
	buffer := new(bytes.Buffer)

	t, err := ttemplate.New("welcome.txt").ParseFiles("templates/emails/welcome.txt")
	if err != nil {
		log.Printf("[emails.go] ParseFIles err: %s", err)
	}

	err = t.Execute(buffer, ed)
	if err != nil {
		log.Printf("[emails.go] Execute err: %s", err)
	}

	return buffer.String()
}

func SendVerificationEmail(email, link string) {
	ed := &EmailData{}
	ed.VerificationLink = link
	ed.EmailMessage = &EmailMessage{
		Subject: "Welcome to PicoStats",
		To:      email, FromName: "PicoStats",
		FromEmail: DEFAULT_EMAIL_SENDER,
		HTML:      ParseHTML(ed),
		Text:      ParseText(ed),
	}

	em.Queue(ed)
}
