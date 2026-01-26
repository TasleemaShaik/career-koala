FROM golang:1.24.4 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/career-koala .

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /out/career-koala /app/career-koala
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/career-koala"]
