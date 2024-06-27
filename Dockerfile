# sudo docker build --rm -t player:latest .
# docker run --init --rm --name player -e USER_PASSWORD="1234qwer" -e PROGRAM_PORT=":5000" -e DEBUG="false" -e WEBSITE_TITLE="Online Player" -e REDIS_ADDRESS="127.0.0.1:6379" -e REDIS_CHAT_EXP="60" -e USER_PASSWORD="123" -e JWT_SECRET="mvS9t48yq5Y32rionewt8yerlgn" -e JWT_EXPIRE_TIME="3" -p 5000:5000 player:latest
# sudo docker stop player
FROM golang:1.22.1-alpine3.19 as build
WORKDIR /app
COPY ./go.mod .
COPY ./go.sum .
RUN export GOPROXY=https://goproxy.cn,direct
RUN go mod download
COPY . /app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -buildvcs=false -trimpath -ldflags="-w -s" -o ./player ./main.go

FROM redis:7.2.5-alpine
WORKDIR /app
COPY --from=build /app/player /app
COPY .env /app/
COPY start.sh /app/
RUN mkdir -p /app/internal/static/video
COPY internal/static internal/static
COPY internal/views internal/views
RUN chmod +x /app/start.sh
EXPOSE 5000
ENTRYPOINT ["/app/start.sh"]