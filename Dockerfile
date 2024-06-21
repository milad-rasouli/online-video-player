# sudo docker build -t player:latest .
# sudo docker run --init --rm --name player -p 5000:5000 player:latest
# sudo docker stop player
FROM golang:1.22.1-alpine3.19 as build
WORKDIR /app
COPY ./go.mod .
COPY ./go.sum .
RUN export GOPROXY=https://goproxy.io,direct
RUN go mod download
COPY . /app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -buildvcs=false -trimpath -ldflags="-w -s" -o ./player ./main.go

FROM redis:7.2.5-alpine
WORKDIR /app
COPY --from=build /app/player /app
COPY .env /app/
COPY start.sh /app/
RUN mkdir internal
COPY internal/static internal/static
COPY internal/views internal/views
RUN chmod +x /app/start.sh
EXPOSE 5000
ENTRYPOINT ["/app/start.sh"]