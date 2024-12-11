# Use an official Go runtime as a parent image
FROM golang:alpine

# Set the working directory inside the container
WORKDIR /go/src/app

# Copy the go.mod and go.sum files to leverage Docker cache
COPY go.mod go.sum ./

# Download module dependencies
RUN go mod tidy

# Copy the local package files to the container's workspace
COPY . .

# Build the Go application
RUN go build -o triaging-sample-app .

# Expose the port on which the application will run
EXPOSE 3000

# Define the command to run the application
CMD ["./triaging-sample-app"]
