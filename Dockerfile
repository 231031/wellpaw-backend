FROM golang:1.25.4-alpine as base

WORKDIR /app/go

# Install necessary tools
RUN apk update && apk add --no-cache git curl

# Install air using Go
RUN go install github.com/air-verse/air@v1.61.0

COPY . .

RUN go mod download

# Create the public/uploads folder
RUN mkdir -p public/uploads

# Expose the port the application will run on
EXPOSE 50001

# Run
CMD ["air", "-c", ".air.toml"]