FROM golang:1.22.2

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -C cmd/jard/ -v -o /app/jard

CMD ["jard"]

