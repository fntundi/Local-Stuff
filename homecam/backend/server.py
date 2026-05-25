"""
Sentinel NOC - Go Backend Proxy Launcher
This FastAPI app acts as a proxy to launch and manage the Go backend.
The actual API is served by the Go binary.
"""
import os
import subprocess
import signal
import sys
import time
import threading
from contextlib import asynccontextmanager
import httpx

# Global process reference
go_process = None

def start_go_backend():
    """Start the Go backend process"""
    global go_process
    
    backend_dir = os.path.dirname(os.path.abspath(__file__))
    go_binary = os.path.join(backend_dir, 'sentinel-noc')
    
    # Build if not exists
    if not os.path.exists(go_binary):
        print("Building Go backend...")
        result = subprocess.run(
            ['/usr/local/go/bin/go', 'build', '-o', 'sentinel-noc', './cmd/api'],
            cwd=backend_dir,
            capture_output=True,
            text=True
        )
        if result.returncode != 0:
            print(f"Build failed: {result.stderr}")
            return None
    
    # Set environment
    env = os.environ.copy()
    env['PORT'] = '8002'  # Go runs on 8002, we proxy from 8001
    env.setdefault('MONGO_URL', 'mongodb://localhost:27017')
    env.setdefault('DB_NAME', 'sentinel_noc')
    
    # Start the Go process
    print(f"Starting Go backend on port 8002...")
    go_process = subprocess.Popen(
        [go_binary],
        cwd=backend_dir,
        env=env,
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
    )
    
    # Log output in background
    def log_output():
        for line in go_process.stdout:
            print(f"[GO] {line.decode().strip()}")
    
    threading.Thread(target=log_output, daemon=True).start()
    
    # Wait for Go backend to be ready
    time.sleep(2)
    return go_process

def stop_go_backend():
    """Stop the Go backend process"""
    global go_process
    if go_process:
        print("Stopping Go backend...")
        go_process.terminate()
        go_process.wait(timeout=10)
        go_process = None

# FastAPI app with lifespan
from fastapi import FastAPI, Request, Response
from fastapi.middleware.cors import CORSMiddleware

@asynccontextmanager
async def lifespan(app: FastAPI):
    # Startup
    start_go_backend()
    yield
    # Shutdown
    stop_go_backend()

app = FastAPI(lifespan=lifespan)

# CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Proxy all requests to Go backend
@app.api_route("/{path:path}", methods=["GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"])
async def proxy(request: Request, path: str):
    async with httpx.AsyncClient() as client:
        # Build target URL
        target_url = f"http://localhost:8002/{path}"
        if request.query_params:
            target_url += f"?{request.query_params}"
        
        # Get request body
        body = await request.body()
        
        # Forward request
        try:
            response = await client.request(
                method=request.method,
                url=target_url,
                headers={k: v for k, v in request.headers.items() if k.lower() not in ['host', 'content-length']},
                content=body,
                timeout=30.0
            )
            
            return Response(
                content=response.content,
                status_code=response.status_code,
                headers=dict(response.headers),
            )
        except httpx.ConnectError:
            return Response(
                content='{"detail": "Backend not available"}',
                status_code=503,
                media_type="application/json"
            )
