FROM golang:1.20-alpine AS build_base
RUN apk add bash ca-certificates git gcc g++ libc-dev
WORKDIR /go/src/github.com/fairDataSociety/FaVe
ENV GO111MODULE=on
COPY go.mod .
COPY go.sum .
RUN go mod download

FROM build_base AS server_builder
ARG TARGETARCH
ARG GITHASH="unknown"
ARG EXTRA_BUILD_ARGS=""
COPY . .
RUN GOOS=linux GOARCH=$TARGETARCH go build $EXTRA_BUILD_ARGS \
      -ldflags '-w -extldflags "-static" ' \
      -o /fave ./cmd/fave-server

FROM alpine AS fave
COPY --from=server_builder /fave /bin/fave
RUN apk add --no-cache --upgrade ca-certificates openssl
ENTRYPOINT ["/bin/fave"]
