# Fly.io Secrets Setup

Run these commands in your terminal to set the production environment variables for your applications. Replace the placeholder values with your actual production credentials.

## API Application (`omni-rapture-api`)

```bash
fly secrets set \
  DATABASE_URL="your_db_url" \
  REDIS_URL="your_redis_url" \
  GEMINI_API_KEY="your_key" \
  ENCRYPTION_KEY="your_base64_32byte_key" \
  JWT_SECRET="your_jwt_secret" \
  TELEGRAM_BOT_TOKEN="your_token" \
  TELEGRAM_WEBHOOK_SECRET="your_secret" \
  R2_ACCOUNT_ID="your_id" \
  R2_BUCKET_NAME="your_bucket" \
  R2_ACCESS_KEY="your_access_key" \
  R2_SECRET_KEY="your_secret_key" \
  GOOGLE_CLIENT_ID="your_id" \
  GOOGLE_CLIENT_SECRET="your_secret" \
  GOOGLE_REDIRECT_URL="your_url" \
  -a omni-rapture-api
```

## Worker Application (`omni-rapture-worker`)

```bash
fly secrets set \
  DATABASE_URL="your_db_url" \
  REDIS_URL="your_redis_url" \
  GEMINI_API_KEY="your_key" \
  ENCRYPTION_KEY="your_base64_32byte_key" \
  -a omni-rapture-worker
```

---

### Deployment Commands

After setting the secrets, use the Makefile to deploy:

```bash
make deploy-api
make deploy-worker
```
