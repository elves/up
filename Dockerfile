FROM golang:1.21.6-alpine3.19

# Runtime dependencies for the app
RUN apk --no-cache add git make rsync zip sqlite python3 py3-beautifulsoup4

# Build app
COPY app /app
RUN go build -o /app/up /app/up.go

# Set up user and directories.
RUN adduser -D -g '' builder
RUN mkdir /data && chown -R builder /data /go
USER builder

CMD ["/app/up", \
     "-secret", "/data/secret", \
     "-master-hook", "/app/master-hook", \
     "-tag-hook", "/app/tag-hook", \
     "-addr", ":8000"]

EXPOSE 8000
