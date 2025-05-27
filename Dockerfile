FROM golang:1.24 AS build-stage

WORKDIR /LoginCenter

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN ls -la
RUN go build -o Server server.go


# Deploy the application binary into a lean image
FROM debian AS build-release-stage

WORKDIR /

COPY --from=build-stage /LoginCenter/Server /Server

EXPOSE 8888

USER nobody

ENTRYPOINT ["/Server"]