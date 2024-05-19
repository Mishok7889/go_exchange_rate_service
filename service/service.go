package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/robfig/cron/v3"
	"gopkg.in/gomail.v2"
	"log"
	"net/http"

	"exchange-service/proto"
)

type ExchangeRateServiceServer struct {
	proto.UnimplementedExchangeRateServiceServer
	rate          float32
	emails        []string
	conn          *pgx.Conn
	smtpPort      int
	emailAddress  string
	emailPassword string
	smtpAddress   string
	cronScheduler *cron.Cron
	entryID       cron.EntryID
}

func NewExchangeRateServiceServer(emailAddress, emailPassword, smtpAddress string, emailPort int) *ExchangeRateServiceServer {
	conn, err := connectDB()
	if err != nil {
		log.Fatal(err)
	}
	err = migrateDB(conn)
	if err != nil {
		log.Fatal(err)
	}
	return &ExchangeRateServiceServer{
		rate:          27.0, // Initial rate, should be updated from an external API
		emails:        []string{},
		conn:          conn,
		emailAddress:  emailAddress,
		emailPassword: emailPassword,
		smtpPort:      emailPort,
		smtpAddress:   smtpAddress,
	}
}

func (s *ExchangeRateServiceServer) GetCurrentRate(ctx context.Context, req *proto.GetCurrentRateRequest) (*proto.GetCurrentRateResponse, error) {
	rate, err := s.fetchCurrentRate()
	if err != nil {
		return nil, err
	}
	return &proto.GetCurrentRateResponse{Rate: rate}, nil
}

func (s *ExchangeRateServiceServer) fetchCurrentRate() (float32, error) {
	resp, err := http.Get("https://api.exchangerate-api.com/v4/latest/USD") // Replace with your chosen API
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}

	rate, ok := data["rates"].(map[string]interface{})["UAH"].(float64)
	if !ok {
		return 0, errors.New("failed to get UAH rate")
	}

	return float32(rate), nil
}

func (s *ExchangeRateServiceServer) SubscribeEmail(ctx context.Context, req *proto.SubscribeEmailRequest) (*proto.SubscribeEmailResponse, error) {

	var existing string
	err := s.conn.QueryRow(ctx, "SELECT email FROM subscriptions WHERE email = $1", req.Email).Scan(&existing)
	if err != nil && err != pgx.ErrNoRows {
		return nil, errors.New("database error")
	}
	if existing != "" {
		return nil, errors.New("Email already subscribed")
	}

	_, err = s.conn.Exec(ctx, "INSERT INTO subscriptions (email) VALUES ($1)", req.Email)
	if err != nil {
		return nil, errors.New("failed to subscribe email")
	}

	return &proto.SubscribeEmailResponse{Message: "E-mail додано"}, nil
}

func (s *ExchangeRateServiceServer) sendEmails() {
	rate, err := s.fetchCurrentRate()
	if err != nil {
		log.Println("Failed to fetch current exchange rate:", err)
		return
	}

	rows, err := s.conn.Query(context.Background(), "SELECT email FROM subscriptions")
	if err != nil {
		log.Println("Failed to fetch emails:", err)
		return
	}
	defer rows.Close()

	emails := []string{}
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			log.Println("Failed to scan email:", err)
			continue
		}
		emails = append(emails, email)
	}

	if len(emails) == 0 {
		log.Println("No emails to send")
		return
	}

	m := gomail.NewMessage()
	m.SetHeader("From", s.emailAddress)
	m.SetHeader("Subject", "Daily Exchange Rate Update")
	m.SetBody("text/plain", "The current exchange rate is: "+fmt.Sprintf("%f", rate))

	d := gomail.NewDialer(s.smtpAddress, s.smtpPort, s.emailAddress, s.emailPassword)

	// Attempt to send emails
	for _, email := range emails {
		m.SetHeader("To", email)
		if err := d.DialAndSend(m); err != nil {
			log.Println("Could not send email to", email, ":", err)
		} else {
			log.Println("Email sent to", email)
		}
	}
}

func (s *ExchangeRateServiceServer) StartScheduler(hour, minute int) {
	s.cronScheduler = cron.New()
	schedule := fmt.Sprintf("%d %d * * *", minute, hour)
	entryID, err := s.cronScheduler.AddFunc(schedule, func() {
		s.sendEmails()
		// Print the next scheduled execution time after sending emails
		entry := s.cronScheduler.Entry(s.entryID)
		log.Println("Next scheduled execution at:", entry.Next)
	})
	if err != nil {
		log.Fatal("Failed to schedule email task:", err)
	}

	s.entryID = entryID
	s.cronScheduler.Start()
	entry := s.cronScheduler.Entry(entryID)
	log.Println("Scheduler started with first execution at:", entry.Next)
}
