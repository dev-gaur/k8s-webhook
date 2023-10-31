FROM golang:1.19

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
COPY *.go ./
COPY internal/ ./

RUN go mod download && go mod tidy

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/engine/reference/builder/#copy

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /mutation-server

EXPOSE 8080

# Run
CMD ["/mutation-server"]