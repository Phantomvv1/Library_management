FROM golang:1.24

WORKDIR /api

COPY . .

RUN go mod tidy

CMD [ "go", "run", "main.go" ]
