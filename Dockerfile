FROM golang:1.13

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

ADD . /app
WORKDIR /app
RUN go get \
 && go build -o packer-provisioner-goss

FROM hashicorp/packer:light
COPY --from=0 /app/packer-provisioner-goss /bin/packer-provisioner-goss
