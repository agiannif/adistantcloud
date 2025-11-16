FROM golang:1-alpine AS builder
WORKDIR /tmp/adistantcloud
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-w -s" ./cmd/adistantcloud/

FROM scratch AS runner
COPY --from=builder /tmp/adistantcloud/adistantcloud /
COPY --from=builder /tmp/adistantcloud/web/static /web/static
CMD ["/adistantcloud"]
