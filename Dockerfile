<<<<<<< Updated upstream
# Stage 1: Build Angular App
FROM node:20 AS frontend

WORKDIR /app/frontend
COPY competition-frontend/package*.json ./

RUN npm install
COPY competition-frontend ./
RUN npm run build --configuration=production

# Stage 2: Build Go backend + embed static
FROM golang:1.24-alpine AS builder

WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/. ./

# Copy Angular build into ./static
COPY --from=frontend /app/frontend/dist/competition-frontend ./static

# Build the Go binary
RUN go build -o server

# Stage 3: Final runtime image
FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /root/
# Copy the server binary
COPY --from=builder /app/backend/server ./server
# Copy the static folder 
COPY --from=builder /app/backend/static ./static

EXPOSE 8080

CMD ["./server"]
=======
#Stage 1: Build Angular App
FROM node:20 as frontend

WORKDIR /app/frontend

#Install dependencies and build Angular app
COPY ./competition-frontend/package*.json ./
RUN npm install
COPY ./competition-frontend ./
RUN npm run build --prod

#Stage 2: Build Go backend
FROM golang:1.24-alpine as backend

WORKDIR /app/backend

COPY ./backend/go.mod ./backend/go.sum ./
RUN go mod tidy
COPY ./backend ./

RUN go build -o server

#Stage 3: Serve Angular with Nginx and run Go backend
FROM nginx:alpine

COPY --from=frontend /app/frontend/dist/competition-frontend/browser /usr/share/nginx/html

COPY ./nginx.conf /etc/nginx/nginx.conf

COPY --from=backend /app/backend/server /server

EXPOSE 80 8080

CMD ["/bin/sh", "-c", "/server & nginx -g 'daemon off;'"]
>>>>>>> Stashed changes
