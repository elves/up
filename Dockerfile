FROM xiaq/alpine-go-pandoc

# Install dependencies from APT
RUN apk update && \
    apk add git make rsync

# Build app
COPY app /app
RUN go build -o /app/up /app/up.go && \
    mkdir /data

CMD ["/app/up", \
     "-secret", "/data/secret", \
     "-master-hook", "/app/master-hook", \
     "-tag-hook", "/app/tag-hook"]

EXPOSE 80
