FROM golang:1.19

# Set destination for COPY
WORKDIR /app

RUN apt-get -y update && apt-get install git
# Download Go modules
COPY go.mod go.sum ./
COPY *.go ./
COPY ./internal/ ./internal

RUN go mod download

# Build
RUN go build -x -o /mutation-server

EXPOSE 8443
# Run
ENTRYPOINT ["/mutation-server"]