package partials

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/mail"
	"net/smtp"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

type Mail struct {
	ID      uint32 // Use the message UID for a stable ID
	Subject string
	Sender  string
	Body    string
	Date    time.Time
}

type mailData struct {
	MailBoxes []string
	Mails []Mail
}

// SMTPConfig is now more appropriately named MailboxConfig
// as it's used for both IMAP and SMTP.
type MailboxConfig struct {
	Server   string
	Port     string
	Username string
	Password string
}

type Handler struct {
	Tmpl          *template.Template
	SMTPConfig    MailboxConfig
	IMAPConfig    MailboxConfig
}

func MailHandler(tmpl *template.Template, smtpConfig MailboxConfig, imapConfig MailboxConfig) *Handler {
	return &Handler{
		Tmpl:       tmpl,
		SMTPConfig: smtpConfig,
		IMAPConfig: imapConfig,
	}
}

func (h *Handler) GetMails(w http.ResponseWriter, r *http.Request) {
	mails, boxes, err := h.FetchMails()
	if err != nil {
		log.Printf("Error fetching mails: %v", err)
		http.Error(w, "Failed to fetch emails", http.StatusInternalServerError)
		return
	}

	mailboxNames := make([]string, len(boxes))
	for i, box := range boxes {
		mailboxNames[i] = box.Name
	}

	data := mailData{
		Mails: mails,
		MailBoxes: mailboxNames,
	}

	err = h.Tmpl.ExecuteTemplate(w, "inbox.html", data)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// FetchMails retrieves emails from the IMAP server
func (h *Handler) FetchMails() ([]Mail, []imap.MailboxInfo, error) {
	// Connect to IMAP server
	log.Println("Connecting to IMAP server:", h.IMAPConfig.Server+":"+h.IMAPConfig.Port)
	c, err := client.DialTLS(h.IMAPConfig.Server+":"+h.IMAPConfig.Port, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("IMAP dial error: %w", err)
	}
	defer c.Logout()
	log.Println("Connected to IMAP server")

	// Login
	log.Println("Logging in to IMAP server with username:", h.IMAPConfig.Username, "and password:", h.IMAPConfig.Password)
	if err := c.Login(h.IMAPConfig.Username, h.IMAPConfig.Password); err != nil {
		return nil, nil, fmt.Errorf("IMAP login error: %w", err)
	}
	log.Println("Logged in successfully")

	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)

	go func() {
		done <- c.List("", "*", mailboxes)
	}()

	log.Println("Mailboxes:")
	resultBoxes := []imap.MailboxInfo{}
	for m := range mailboxes {
		log.Printf("- %s (Flags: %v)", m.Name, m.Attributes)
		resultBoxes = append(resultBoxes, *m)
	}

	mbox, err := c.Select("INBOX", false)
	if err != nil {
		return nil, nil, fmt.Errorf("IMAP select INBOX error: %w", err)
	}
	log.Printf("Selected INBOX with %d messages", mbox.Messages)
	// Get the last 10 messages
	if mbox.Messages == 0 {
		return []Mail{}, resultBoxes, nil
	}
	from := uint32(1)
	to := mbox.Messages
	if mbox.Messages > 10 {
		from = mbox.Messages - 9
	}

	seqSet := new(imap.SeqSet)
	seqSet.AddRange(from, to)

	// Get message envelope, body, and UID
	var section imap.BodySectionName
	items := []imap.FetchItem{imap.FetchEnvelope, section.FetchItem(), imap.FetchUid}
	messages := make(chan *imap.Message, 10)

	// The go func() was problematic. It's better to handle the fetch directly.
	if err := c.Fetch(seqSet, items, messages); err != nil {
		return nil, resultBoxes, fmt.Errorf("IMAP fetch error: %w", err)
	}

	var mails []Mail
	for msg := range messages {
		if msg == nil {
			log.Println("Server sent a nil message")
			continue
		}

		r := msg.GetBody(&section)
		if r == nil {
			log.Println("Server didn't return message body")
			continue
		}

		// Use io.ReadAll instead of deprecated ioutil.ReadAll
		bodyBytes, err := io.ReadAll(r)
		if err != nil {
			log.Printf("Error reading message body: %v", err)
			continue // Or handle error appropriately
		}
		
		var senderAddress string
		if len(msg.Envelope.From) > 0 {
			senderAddress = msg.Envelope.From[0].Address()
		}

		mail := Mail{
			ID:      msg.Uid, // Use UID for a persistent identifier
			Subject: msg.Envelope.Subject,
			Sender:  senderAddress,
			Body:    string(bodyBytes),
			Date:    msg.Envelope.Date,
		}

		mails = append(mails, mail)
	}

	// Reverse the slice to show newest emails first
	for i, j := 0, len(mails)-1; i < j; i, j = i+1, j-1 {
		mails[i], mails[j] = mails[j], mails[i]
	}


	return mails, resultBoxes, nil
}

// SendMail sends an email using the standard net/smtp package
func (h *Handler) SendMail(to []string, subject, body string) error {
	// The `emersion/go-smtp` package is for BUILDING an SMTP server.
	// For SENDING mail (as a client), the standard library `net/smtp` is used.

	// Set up authentication information.
	auth := smtp.PlainAuth("", h.SMTPConfig.Username, h.SMTPConfig.Password, h.SMTPConfig.Server)

	// Format the email message according to RFC 822.
	// This is a much more robust way to build the email.
	fromAddr := mail.Address{Name: "", Address: h.SMTPConfig.Username}
	toAddr := mail.Address{Name: "", Address: to[0]} // Assuming one recipient for simplicity

	headers := make(map[string]string)
	headers["From"] = fromAddr.String()
	headers["To"] = toAddr.String()
	headers["Subject"] = subject
	headers["Content-Type"] = `text/plain; charset="UTF-8"`

	var msg bytes.Buffer
	for k, v := range headers {
		msg.WriteString(k + ": " + v + "\r\n")
	}
	msg.WriteString("\r\n")
	msg.WriteString(body)

	// Send the email.
	// This function handles connecting, authenticating, and sending the email.
	addr := h.SMTPConfig.Server + ":" + h.SMTPConfig.Port
	err := smtp.SendMail(addr, auth, fromAddr.Address, to, msg.Bytes())
	if err != nil {
		return fmt.Errorf("smtp.SendMail error: %w", err)
	}

	log.Println("Email sent successfully!")
	return nil
}

// SendMailHandler handles the email sending form submission
func (h *Handler) SendMailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	to := r.FormValue("to")
	subject := r.FormValue("subject")
	body := r.FormValue("body")

	if to == "" || subject == "" || body == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	err := h.SendMail([]string{to}, subject, body)
	if err != nil {
		log.Printf("Error sending email: %v", err)
		http.Error(w, "Failed to send email", http.StatusInternalServerError)
		return
	}

	// Redirect back to the inbox or a success page
	http.Redirect(w, r, "/inbox", http.StatusSeeOther)
}