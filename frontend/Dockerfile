# Build stage
FROM node:18-alpine AS builder

WORKDIR /app

# Hardcode environment variables for build
ENV VITE_CLERK_PUBLISHABLE_KEY=pk_test_bGVnYWwtc2FpbGZpc2gtODQuY2xlcmsuYWNjb3VudHMuZGV2JA
ENV VITE_API_URL=https://golf-league-backend-130681562264.us-central1.run.app

COPY package*.json ./
RUN npm install

COPY . .
RUN npm run build

# Production stage
FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
