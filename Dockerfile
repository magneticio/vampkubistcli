FROM golang:1.11.4

COPY . /go/src/github.com/magneticio/vamp2cli

WORKDIR /go/src/github.com/magneticio/vamp2cli

RUN go get

RUN go build -o bin/vamp2cli \
    && cp bin/vamp2cli /usr/local/bin/vamp2cli \
    && chmod +x /usr/local/bin/vamp2cli

RUN useradd -ms /bin/bash vamp
USER vamp
WORKDIR /home/vamp
CMD /bin/bash
