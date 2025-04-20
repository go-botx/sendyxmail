FROM golang:alpine AS builder
RUN apk update --no-cache && apk add --no-cache ca-certificates
WORKDIR /build
ADD go.mod .
ADD go.sum .
RUN go mod download
COPY . .
RUN go build -o sendyxmail .

FROM alpine
COPY docker/docker-entrypoint.sh /
WORKDIR /opt/bin
RUN \
  apk add --no-cache ca-certificates && \
  update-ca-certificates && \
  chmod +x /docker-entrypoint.sh && \
  mkdir -p /mutes && \
  chmod a+w /mutes 
COPY --from=builder /build/sendyxmail .
EXPOSE 8000
VOLUME [ "/mutes", "/certs" ]
ENTRYPOINT [ "/docker-entrypoint.sh" ]
CMD ["./sendyxmail"]