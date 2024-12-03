# FASTAPI Helloworld Server

from fastapi import FastAPI, Request
import os
import uvicorn
import random
import time

PORT = os.getenv("PORT", 8000)
app = FastAPI()


@app.get("/")
async def root(request: Request):
    start_time = time.time()
    # sleep 200 - 500 ms randomly
    sleep_time = random.randint(200, 500) / 1000
    time.sleep(sleep_time)
    end_time = time.time()
    return {"message": f"Hello World from port {PORT}", "duration": round((end_time - start_time) * 1000)}


if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=int(PORT))
