FROM golang:1.22.5-alpine3.19 as builder

ADD . /go/src/app
WORKDIR /go/src/app

COPY . .

RUN go mod download
RUN go mod tidy
RUN go build -o=mem-db cmd/main.go

FROM scratch

COPY --from=builder /go/src/app/mem-db /app/mem-db

# Set the working directory
WORKDIR /app

# Command to run the binary
CMD ["./mem-db"]
