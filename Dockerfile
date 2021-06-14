FROM pandoc/core:latest as pandoc
FROM golang:1-alpine
COPY --from=pandoc /usr/local/bin/pandoc /usr/local/bin/pandoc
# Runtime dependencies for pandoc.
RUN apk --no-cache add gmp libffi lua5.3 lua5.3-lpeg

# Runtime dependencies for the app
RUN apk --no-cache add git make rsync zip python3 py3-pip
RUN pip3 install beautifulsoup4

# Build app
COPY app /app
RUN go build -o /app/up /app/up.go

# Set up user and data directory.
RUN adduser -D -g '' builder
RUN mkdir /data && chown builder /data
USER builder

CMD ["/app/up", \
     "-secret", "/data/secret", \
     "-master-hook", "/app/master-hook", \
     "-tag-hook", "/app/tag-hook", \
     "-addr", ":8000"]

EXPOSE 8000
