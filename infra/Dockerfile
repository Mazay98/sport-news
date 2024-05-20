FROM golang:1.22.0-alpine AS build
RUN apk add --no-cache make
RUN apk add --no-cache build-base
WORKDIR /opt/sport-news
COPY . .
ARG ACTION
RUN make $ACTION

FROM alpine:3.18
WORKDIR /opt/sport-news
RUN apk add --no-cache tzdata
COPY --from=build /opt/sport-news/bin/sport-news .

CMD ["./sport-news"]
