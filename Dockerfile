FROM xiaq/alpine-go-pandoc

# Install dependencies from APT
RUN apk update && \
    apk add git make rsync zip

# Build app
COPY app /app
RUN go build -o /app/up /app/up.go && \
    mkdir /data

# Set up data directory and permission
RUN adduser -D -g '' appuser
RUN mkdir /data && chown appuser /data
USER appuser

CMD ["/app/up", \
     "-secret", "/data/secret", \
     "-master-hook", "/app/master-hook", \
     "-tag-hook", "/app/tag-hook", \
     "-addr", ":8000"]

EXPOSE 8000
