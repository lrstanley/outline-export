FROM golang:latest as build
WORKDIR /build

COPY . /build/
WORKDIR /build
RUN \
	--mount=type=cache,target=/root/.cache \
	--mount=type=cache,target=/go \
	make build

FROM alpine:3.22
RUN apk add --no-cache ca-certificates
COPY --from=build /build/outline-export /usr/local/bin/outline-export

WORKDIR /
ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
CMD ["/usr/local/bin/outline-export"]
