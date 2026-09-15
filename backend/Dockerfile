FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /lingolift .

FROM alpine:3.21
RUN addgroup -S app && adduser -S app -G app
COPY --from=build /lingolift /usr/local/bin/lingolift
USER app
EXPOSE 8080
ENV PORT=8080
CMD ["lingolift"]
