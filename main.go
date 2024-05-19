package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"exchange-service/proto"
	"exchange-service/service"
	"google.golang.org/grpc"
)

type ExchangeRate struct {
	Rate float32 `json:"rate"`
}

func main() {
	godotenv.Load()

	emailAddress := os.Getenv("EMAIL_ADDRESS")
	emailPassword := os.Getenv("EMAIL_PASSWORD")
	smtpAddress := os.Getenv("SMTP_ADDRESS")
	emailPort, err := strconv.Atoi(os.Getenv("EMAIL_SMTP_PORT"))

	lis, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	srv := service.NewExchangeRateServiceServer(emailAddress, emailPassword, smtpAddress, emailPort)
	proto.RegisterExchangeRateServiceServer(grpcServer, srv)

	go func() {
		log.Printf("gRPC server listening at %v", lis.Addr())
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Start the scheduler
	startTime := time.Now().Add(1 * time.Minute)
	go srv.StartScheduler(startTime.Hour(), startTime.Minute())

	r := mux.NewRouter()
	r.HandleFunc("/rate", func(w http.ResponseWriter, r *http.Request) {
		req := &proto.GetCurrentRateRequest{}
		res, err := srv.GetCurrentRate(r.Context(), req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		exchangeRate := ExchangeRate{Rate: res.Rate}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(exchangeRate)
	}).Methods("GET")
	r.HandleFunc("/subscribe", func(w http.ResponseWriter, r *http.Request) {
		email := r.FormValue("email")
		if email == "" {
			http.Error(w, "Email is required", http.StatusBadRequest)
			return
		}

		req := &proto.SubscribeEmailRequest{Email: email}
		res, err := srv.SubscribeEmail(r.Context(), req)
		if err != nil {
			if err.Error() == "Email already subscribed" {
				http.Error(w, err.Error(), http.StatusConflict)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(res.Message))
	}).Methods("POST")

	http.Handle("/", r)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("HTTP server listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
