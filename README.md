# Exchange Rate Service

This service provides an API to get the current exchange rate of USD to UAH and to subscribe to daily email updates of the exchange rate. The service is built using Go and leverages gRPC for communication and a PostgreSQL database for storing subscriptions. The email credentials for sending updates are read from an `.env` file, while the database address is currently hardcoded. The web API runs on port 8080 and the gRPC service runs on port 8081, both of which are also hardcoded.

## Features

- Get the current USD to UAH exchange rate.
- Subscribe an email address to receive daily exchange rate updates.
- Send daily emails to all subscribed email addresses with the current exchange rate.
- Cron-based scheduling to send emails at a specified time each day.

## Configuration

Create a `.env` file in the root of the project with the following structure, filling in the required values:

```
EMAIL_ADDRESS=your-email@example.com  
EMAIL_PASSWORD=your-email-password  
SMTP_ADDRESS=smtp.example.com  
EMAIL_SMTP_PORT=587`
```


## Running the Application

### Prerequisites

- Docker
- Docker Compose

### Running with Docker

1. **Build the Docker image:**

   ```sh
   docker build -t exchange-rate-service .
   
2. **Run the Docker container using Docker Compose:**

   ```sh
   docker-compose up
   
This will start the application with the web API available on port <b>8080</b> and the gRPC service on port <b>8081</b>.

### Postman Collection
A Postman collection is provided for testing the API endpoints. This collection includes requests for all available endpoints and can be used to quickly verify the functionality of the service.

### Notes

- The api schema is located in the gses2swaggger.yaml file.
- The PostgreSQL database address is currently hardcoded in the source code.
- Email credentials are read from the .env file and must be provided for the service to send email updates.
- The web API runs on port 8080 and the gRPC service runs on port 8081, both of which are hardcoded.

### Future Improvements

- Add unit and integration tests with appropriate mocking.
- Make database and port configurations configurable through environment variables.
- Improve error handling and logging throughout the application.