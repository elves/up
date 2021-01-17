FROM pandoc/core:latest as pandoc
FROM golang:1-alpine
COPY --from=pandoc /usr/local/bin/pandoc /usr/local/bin/pandoc
# Runtime dependencies for pandoc.
RUN apk --no-cache add gmp libffi lua5.3 lua5.3-lpeg

# Runtime dependencies for the app
RUN apk --no-cache add git make rsync zip

# Build app
COPY app /app
RUN go build -o /app/up /app/up.go

# Set up data directory and permission. The user is called travis to make it
# easier to emulate the GOROOT and GOPATH of our Travis builds.
RUN adduser -D -g '' travis
RUN mkdir /data && chown travis /data
RUN ln -s /usr/local/go /home/travis/goroot
USER travis

CMD ["/app/up", \
     "-secret", "/data/secret", \
     "-master-hook", "/app/master-hook", \
     "-tag-hook", "/app/tag-hook", \
     "-addr", ":8000"]

EXPOSE 8000
