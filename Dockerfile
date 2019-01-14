FROM golang:1.11.4

COPY . /go/src/github.com/magneticio/vamp2cli

WORKDIR /go/src/github.com/magneticio/vamp2cli

RUN go get

RUN go build -o bin/$AppName

FROM ubuntu:16.04

COPY --from=0 /go/src/github.com/magneticio/vamp2cli/bin/$AppName /usr/local/bin/$AppName
RUN chmod +x /usr/local/bin/$AppName

RUN apt-get update && apt-get install -y apt-transport-https curl wget && \
    curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add - && \
    echo "deb https://apt.kubernetes.io/ kubernetes-xenial main" | tee -a /etc/apt/sources.list.d/kubernetes.list && \
    apt-get update && \
    apt-get install -y kubectl

RUN useradd -ms /bin/bash vamp
USER vamp
WORKDIR /home/vamp
CMD /bin/bash
