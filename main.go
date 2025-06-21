package main

import (
	pages "hmdogan00/email-client/pages"
	partials "hmdogan00/email-client/partials"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

var tmpl *template.Template

func getMailConfig() (partials.MailboxConfig, partials.MailboxConfig) {
	imapConfig := partials.MailboxConfig{
		Server:   "imap.metu.edu.tr",
		Port:     "993",
		Username: os.Getenv("OUTLOOK_EMAIL"),
		Password: os.Getenv("OUTLOOK_APP_PASSWORD"),
	}

	smtpConfig := partials.MailboxConfig{
		Server:   "mail.metu.edu.tr",
		Port:     "465",
		Username: os.Getenv("OUTLOOK_EMAIL"),
		Password: os.Getenv("OUTLOOK_APP_PASSWORD"),
	}

	return imapConfig, smtpConfig
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}
	// parse partial templates
	tmpl = template.Must(template.ParseGlob("templates/partials/*.html"))
	// Initialize handlers
	timeHandler := partials.TimeHandler(tmpl)

	imapConfig, smtpConfig := getMailConfig()
	mailHandler := partials.MailHandler(tmpl, smtpConfig, imapConfig)

	// Define HTTP handlers
	http.HandleFunc("/", pages.Index)
	http.HandleFunc("/mails", pages.Mails)
	http.HandleFunc("/get-time", timeHandler.GetTime)
	http.HandleFunc("/mails/get", mailHandler.GetMails)

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Println("Server starting on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
