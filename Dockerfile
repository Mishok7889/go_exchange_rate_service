# Step 1: Use the official Golang 1.16 image to create a build stage
FROM golang:1.22.3-alpine AS builder

# Step 2: Set the Current Working Directory inside the container
WORKDIR /app

# Step 3: Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Step 4: Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Step 5: Copy the source code into the container
COPY . .

# Step 6: Tidy the module, ensure dependencies are correct
RUN go mod tidy

# Step 7: Build the Go app
RUN go build -o main .

# Step 8: Use a minimal Docker image to serve the app
FROM alpine:latest

# Step 9: Install necessary certificates for HTTPS
RUN apk update && apk add --no-cache ca-certificates

# Step 10: Set the Current Working Directory inside the container
WORKDIR /root/

# Step 11: Copy the pre-built binary file from the previous stage
COPY --from=builder /app/main .

# Step 12: Run the binary program
CMD ["./main"]
