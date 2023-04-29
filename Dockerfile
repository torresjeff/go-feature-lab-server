FROM golang:1.20.3-alpine AS build

RUN mkdir /src
ADD . /src
WORKDIR /src

RUN go build -o /tmp/featurelab ./app.go


FROM alpine:edge
COPY --from=build /tmp/featurelab /sbin/featurelab

CMD /sbin/featurelab