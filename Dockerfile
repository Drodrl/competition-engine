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