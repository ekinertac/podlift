from fastapi import FastAPI
import os
import time

app = FastAPI(title="podlift FastAPI Example - v2")
startup_time = time.time()

@app.get("/")
async def root():
    return {"message": "Version 2 deployed!", "version": os.getenv("APP_VERSION", "v2")}

@app.get("/health")
async def health():
    return {"status": "healthy", "uptime": int(time.time() - startup_time)}

@app.get("/api/users")
async def get_users():
    return {"users": [{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}]}

@app.get("/api/info")
async def get_info():
    return {"environment": os.getenv("ENVIRONMENT", "production"), "version": "2.0"}
