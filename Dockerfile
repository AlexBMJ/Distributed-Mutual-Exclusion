FROM golang:1.17

ENV NODE_NAME ""
ENV ADVERTISE_ADDRESS "127.0.0.1"
ENV CLUSTER_ADDRESS "127.0.0.1"

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

ENTRYPOINT ["go", "run", "./node"]