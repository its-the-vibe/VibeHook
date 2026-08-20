FROM --platform=$BUILDPLATFORM golang:1.27.0-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath -ldflags='-s -w' -o /out/vibehook ./cmd/vibehook

FROM gcr.io/distroless/static-debian13:nonroot

WORKDIR /app

COPY --from=builder /out/vibehook /app/vibehook

USER nonroot:nonroot

ENTRYPOINT ["/app/vibehook"]
