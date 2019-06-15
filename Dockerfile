FROM dockercore/golang-cross:1.12.3 AS build-env
ADD . /root/src
RUN cd /root/src && CGO_ENABLED=0 go build -o vamp

FROM alpine:3.9
WORKDIR /app
COPY --from=build-env /root/src/vamp /app/
ENTRYPOINT ["./vamp", "adapterservice", "--port", "9000", "-v"]

EXPOSE 9000
