FROM golang:1.15

WORKDIR /build

COPY . .

RUN go build ./main.go

EXPOSE 8080

CMD ./main