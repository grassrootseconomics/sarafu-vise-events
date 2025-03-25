FROM golang:1.23.4-bookworm AS build

ENV CGO_ENABLED=1

ARG BUILDPLATFORM
ARG TARGETPLATFORM
ARG BUILD=dev

WORKDIR /build
COPY . .
RUN apt-get update && apt-get install -y --no-install-recommends \
    libgdbm-dev \
    git \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /build
RUN echo "Building on $BUILDPLATFORM, building for $TARGETPLATFORM"
RUN go mod download
RUN go build -tags logwarn -o sarafu-vise-events -ldflags="-X main.build=${BUILD} -s -w" cmd/main.go

FROM debian:bookworm-slim

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
    libgdbm-dev \
    ca-certificates \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /service

COPY --from=build /build/sarafu-vise-events .
COPY --from=build /build/LICENSE .
COPY --from=build /build/.env.example .
RUN mv .env.example .env

CMD ["./sarafu-vise-events"]