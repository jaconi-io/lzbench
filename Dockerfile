FROM debian:11-slim as builder

RUN apt update \
 && apt install --yes --no-install-recommends g++ gcc libc-dev make \
 && rm --recursive --force /var/lib/apt/lists/*

COPY . .

RUN make

FROM debian:11-slim

COPY --from=builder /lzbench /usr/bin/lzbench

ENTRYPOINT [ "lzbench" ]
