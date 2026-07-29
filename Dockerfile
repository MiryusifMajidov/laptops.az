# syntax=docker/dockerfile:1
# Laptops.az — tək image: Go backend həm API-ni, həm admin paneli, həm müştəri saytını verir.

# ---- 1) Frontend-ləri build et (admin + sayt) ----
FROM node:20-alpine AS web
WORKDIR /web

# admin panel (Vite base '/admin/')
COPY frontend/package.json frontend/package-lock.json frontend/
RUN cd frontend && npm ci
COPY frontend/ frontend/
RUN cd frontend && npm run build

# müştəri saytı (Vite base '/')
COPY website/package.json website/package-lock.json website/
RUN cd website && npm ci
COPY website/ website/
RUN cd website && npm run build

# ---- 2) Go binary-ni build et (frontend build-ləri embed olunur) ----
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
# embed üçün build-ləri yerləşdir: static_prod.go bunları binary-ə qatır
COPY --from=web /web/frontend/dist ./web/admin
COPY --from=web /web/website/dist ./web/site
RUN CGO_ENABLED=0 go build -tags prod -trimpath -ldflags "-s -w" -o /out/laptops-backend .

# ---- 3) Runtime (kiçik image) ----
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /out/laptops-backend /app/laptops-backend
# ilk boot-da kalıcı volume-a köçürüləcək seed data (real baza + şəkillər)
COPY backend/laptops.db /app/seed/laptops.db
COPY backend/uploads/ /app/seed/uploads/
# DATA_DIR = kalıcı volume (fly.toml-dakı mount); SEED_DIR = image-dəki ilkin data
ENV PORT=8080 DATA_DIR=/data SEED_DIR=/app/seed TZ=Asia/Baku
EXPOSE 8080
ENTRYPOINT ["/app/laptops-backend"]
