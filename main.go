package main

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"log"
	"net/smtp"
	"net/textproto"
	"os"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/jordan-wright/email"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	dbFile := "mail.db"
	if len(os.Args) > 1 {
		dbFile = os.Args[1]
	}
	db, err := initDB(dbFile)
	if err != nil {
		log.Fatal(err)
	}
	err = loadConfig(db)
	if err != nil {
		log.Fatal(err)
	}
	dests, err := emails(db)
	if err != nil {
		log.Fatal(err)
	}
	for _, d := range dests {
		fmt.Printf("Sending to '%v <%v>' ..", d.fullName, d.email)
		anInvite := invite(d.fullName, d.email)
		//fmt.Println(anInvite)
		err := send(d.email, d.firstName, d.fullName, anInvite)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf(" done\n")
	}

}

type dest struct {
	firstName string
	fullName  string
	email     string
}

var (
	config map[string]string
)

func initDB(dbFile string) (*sql.DB, error) {
	dbType := "sqlite3"
	db, err := sql.Open(dbType, dbFile)
	if err != nil {
		log.Printf("sql.Open(%v, %v) error: %v", dbType, dbFile, err)
		return nil, fmt.Errorf("sql.Open(%v, %v) error: %w", dbType, dbFile, err)
	}
	return db, nil
}

func loadConfig(db *sql.DB) error {
	config = make(map[string]string)
	query := "SELECT name, value FROM configuration"
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("db.Query(%v) error: %v", query, err)
		return fmt.Errorf("db.Query(%v) error: %w", query, err)
	}
	defer rows.Close()
	for rows.Next() {
		var name, value string
		err = rows.Scan(&name, &value)
		if err != nil {
			log.Printf("rows.Scan() error: %v", err)
			return fmt.Errorf("rows.Scan() error: %w", err)
		}
		config[name] = value
	}
	err = rows.Err()
	if err != nil {
		log.Printf("rows.Err() error: %v", err)
		return fmt.Errorf("rows.Err() error: %w", err)
	}
	return nil
}

func emails(db *sql.DB) ([]dest, error) {
	var dests []dest
	query := "SELECT email, first_name, full_name FROM emails"
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("db.Query(%v) error: %v", query, err)
		return nil, fmt.Errorf("db.Query(%v) error: %w", query, err)
	}
	defer rows.Close()
	for rows.Next() {
		var email, firstName, fullName string
		err = rows.Scan(&email, &firstName, &fullName)
		if err != nil {
			log.Printf("rows.Scan() error: %v", err)
			return nil, fmt.Errorf("rows.Scan() error: %w", err)
		}
		dests = append(dests, dest{email: email, firstName: firstName, fullName: fullName})
	}
	err = rows.Err()
	if err != nil {
		log.Printf("rows.Err() error: %v", err)
		return nil, fmt.Errorf("rows.Err() error: %w", err)
	}
	return dests, nil
}

func invite(attendeeName, attendeeEmail string) string {
	if config["UID"] == "" {
		return ""
	}
	dtStart, err := time.Parse(time.RFC3339, config["start"])
	if err != nil {
		log.Fatal(err)
	}
	dtEnd, err := time.Parse(time.RFC3339, config["end"])
	if err != nil {
		log.Fatal(err)
	}

	cal := ics.NewCalendarFor(config["Organization"])
	cal.SetMethod(ics.MethodRequest)
	cal.SetCalscale("GREGORIAN")
	event := cal.AddEvent(config["UID"])
	event.SetStatus(ics.ObjectStatusConfirmed)
	event.SetLocation(config["location"])
	event.SetURL(config["URL"])
	event.SetSummary(config["summary"])
	event.SetDescription(config["description"])
	event.SetStartAt(dtStart)
	event.SetEndAt(dtEnd)
	event.SetOrganizer("mailto:"+config["organizerEmail"], ics.WithCN(config["organizerName"]))
	event.AddAttendee(attendeeEmail, ics.WithCN(attendeeName),
		ics.WithRSVP(true),
		ics.ParticipationStatusNeedsAction,
		ics.ParticipationRoleReqParticipant,
	)
	return cal.Serialize()
}

func send(destEmail, firstName, fullName, anInvite string) error {
	e := email.NewEmail()
	e.From = fmt.Sprintf("%v <%v>", config["fromName"], config["fromEmail"])
	e.To = []string{fmt.Sprintf("%v <%v>", fullName, destEmail)}
	e.Subject = config["subject"]
	spFirstName := fmt.Sprintf(" %v", firstName)
	if firstName == "" {
		spFirstName = ""
	}
	e.Text = []byte(strings.ReplaceAll(config["bodyText"], " [Name]", spFirstName))
	e.HTML = []byte(strings.ReplaceAll(config["bodyHTML"], " [Name]", spFirstName))
	if anInvite != "" {
		invite := &email.Attachment{
			Filename:    "invite.ics",
			ContentType: "text/calendar; charset=utf-8; method=REQUEST; name=invite.ics",
			Header:      textproto.MIMEHeader{},
			Content:     []byte(anInvite),
		}
		e.Attachments = append(e.Attachments, invite)
	}

	err := e.SendWithStartTLS(config["smtpHost"]+":"+config["smtpPort"],
		smtp.PlainAuth("", config["smtpUsername"], config["smtpPassword"], config["smtpHost"]), &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         config["smtpHost"],
		})
	if err != nil {
		log.Printf("e.SendWithStartTLS error: %v", err)
		return fmt.Errorf("e.SendWithStartTLS error: %w", err)
	}
	return err
}
