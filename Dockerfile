FROM golang:1.23.0-alpine3.19

# Runtime dependencies for the app
RUN apk --no-cache add git make rsync zip python3 py3-beautifulsoup4 coreutils
# $GOPATH/bin is in $PATH in the base golang image
RUN go install src.elv.sh/cmd/elvish@v0.21.0

# Build app
COPY app /app
RUN cd /app && go build -o up .

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
