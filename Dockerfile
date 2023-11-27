# syntax=docker/dockerfile:1

# ----- build stage -----
FROM golang:1.25-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 go build \
    -ldflags "-s -w -X github.com/Fzgt/cronflux/internal/buildinfo.Version=${VERSION}" \
    -o /out/cronflux ./cmd/cronflux

# ----- runtime stage -----
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/cronflux /usr/local/bin/cronflux
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/cronflux"]
