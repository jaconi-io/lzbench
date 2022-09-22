FROM debian:11-slim as lzbench

RUN apt update \
 && apt install --yes --no-install-recommends g++ gcc libc-dev make \
 && rm --recursive --force /var/lib/apt/lists/*

COPY . .

RUN make

FROM golang:1.19 as wrapper

WORKDIR /app
COPY . .
ENV CGO_ENABLED=0
RUN go build -o app

FROM debian:11-slim

COPY --from=lzbench /lzbench /usr/bin/lzbench
COPY --from=wrapper /app/app /usr/bin/wrapper

ENTRYPOINT [ "wrapper" ]
