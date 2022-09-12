# syntax=docker/dockerfile:experimental
# ---
FROM golang:1.18 AS build

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

WORKDIR /work
COPY . /work

# Build admission-webhook
RUN --mount=type=cache,target=/root/.cache/go-build,sharing=private \
  go build -o bin/admission-webhook .

# ---
FROM ubuntu AS run
RUN apt update && apt install -y curl # mandatory for public certificates

COPY --from=build /work/bin/admission-webhook /usr/local/bin/
WORKDOR .
COPY key.pub /

CMD ["admission-webhook"]
