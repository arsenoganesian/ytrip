FROM golang:alpine AS builder
RUN mkdir /build
ADD *.go go.mod go.sum /build/
WORKDIR /build
RUN go build -o ytrip .

FROM alpine AS runner
RUN apk -U add ffmpeg yt-dlp
RUN mkdir /app
WORKDIR /app
COPY --from=builder /build/ytrip /app
CMD [ "./ytrip" ]
