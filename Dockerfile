FROM golang:1.26-alpine AS build

WORKDIR /src
RUN apk add --no-cache build-base ca-certificates
COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY networking ./networking
COPY pkg ./pkg
COPY sdk ./sdk
COPY i18n ./i18n

ARG VERSION=v0.0.1
ARG COMMIT=unknown
RUN CGO_ENABLED=1 go build -trimpath -ldflags="-s -w -X github.com/niflaot/pixels/pkg/build.Version=${VERSION} -X github.com/niflaot/pixels/pkg/build.CommitHash=${COMMIT}" -o /out/pixels ./cmd

FROM alpine:3.23

WORKDIR /app

RUN apk add --no-cache ca-certificates libgcc
COPY --from=build /out/pixels /app/pixels
COPY --from=build /src/i18n /app/i18n

ENV PIXELS_ENV=production \
    PIXELS_HOST=0.0.0.0 \
    PIXELS_PORT=3000

EXPOSE 3000

ENTRYPOINT ["/app/pixels"]
