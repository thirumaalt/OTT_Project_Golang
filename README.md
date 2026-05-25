# MyFlix — Go Backend

Go rewrite of the original Java/Python microservices stack.

## Structure

```
Project_Golang/
├── api-gateway/          # Reverse proxy + JWT auth (replaces FastAPI gateway)
├── auth-service/         # Register / Login / JWT generation
├── user-service/         # Users, profiles, watch history, watchlist
├── analytics-service/    # User action tracking
├── recommendation-service/
├── watchhistory-service/
├── payment-service/      # Razorpay order management
├── subscription-service/ # Subscription plans
└── docker-compose.yml
```

## Run locally

```bash
cp ../Project_1_Traditional/.env.example .env
# Set JWT_SECRET in .env

docker compose up --build
```

Gateway is available at `http://localhost:8094`.

## Service ports

| Service              | Port |
|----------------------|------|
| api-gateway          | 8094 |
| auth-service         | 8083 |
| user-service         | 8085 |
| analytics-service    | 8086 |
| recommendation-service | 8087 |
| watchhistory-service | 8090 |
| payment-service      | 8091 |
| subscription-service | 8093 |

## API endpoints (same as original)

### Auth (public)
- `POST /api/auth/register`
- `POST /api/auth/login`

### User (requires Bearer token)
- `POST   /api/user`
- `GET    /api/user/:id`
- `POST   /api/user/profiles`
- `GET    /api/user/profiles?userId=`
- `DELETE /api/user/profiles/:id`
- `POST   /api/user/history`
- `GET    /api/user/history?profileId=`
- `POST   /api/user/watchlist`
- `DELETE /api/user/watchlist?profileId=&mediaPath=`
- `GET    /api/user/watchlist?profileId=`
- `GET    /api/user/trending`

### Analytics
- `POST /api/analytics/record?userId=&action=&contentId=`
- `GET  /api/analytics/all`
- `GET  /api/analytics/stats`

### Watch History
- `POST /api/watch-history/record?userId=&contentId=`
- `GET  /api/watch-history/all`

### Recommendations
- `POST /api/recommendations/add?userId=&content=`
- `GET  /api/recommendations/get?userId=`

### Payment
- `POST /api/payment/create-order`  body: `{"amount": 499}`
- `POST /api/payment/capture-payment?orderId=`
- `GET  /api/payment/status/:orderId`

### Subscription
- `GET  /api/subscription/user/:userId`
- `POST /api/subscription/upgrade?userId=&plan=`

## Health checks
- `GET /health` — gateway liveness
- `GET /system/health` — all services status
- `GET /actuator/health` — per-service health
