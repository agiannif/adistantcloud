FROM golang:1-alpine AS builder
COPY . /tmp/adistantcloud
WORKDIR /tmp/adistantcloud
RUN go build -ldflags "-s -w" ./cmd/adistantcloud/

FROM scratch AS runner
COPY --from=builder /tmp/adistantcloud/adistantcloud /
COPY --from=builder /tmp/adistantcloud/web/static /web/static
CMD ["/adistantcloud"]
