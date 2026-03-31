# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.26 AS build

ARG TARGETARCH
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY mod/*.go ./mod/
COPY bin/*.go ./bin/
# ldflags: -w strips DWARF symbol table | -s removes symbol table and debug info
# gcflags: all=-l disables func inlining
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} \
    go build \
    -ldflags="-w -s" -gcflags="all=-l" \
    -o /docker-gatus-generator \
    ./bin


FROM scratch

WORKDIR /

COPY --from=build /docker-gatus-generator /docker-gatus-generator

USER 65534:65534
ENTRYPOINT ["/docker-gatus-generator"]
