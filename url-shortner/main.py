from typing import Union
import uvicorn
from fastapi import FastAPI, APIRouter, HTTPException, Depends

# import  JSONResponse
from fastapi.responses import JSONResponse

app = FastAPI()

from pydantic import BaseModel, Field
from db import create_pool
from ranges import get_next_id
from encoding import encode

from datetime import timedelta
from datetime import datetime


db_pool = None

class URL(BaseModel):
    short_code : str | None = Field(None, title="Short Code", description="The custom short code for the URL")
    long_url: str = Field(..., title="Long URL", description="The long URL to be shortened")
    expiration_date: datetime | None = Field(None, title="Expiration Date", description="The expiration date for the short URL")

class URLAnalytics(BaseModel):
    short_url: str = Field(..., title="Short URL", description="The shortened URL")
    clicks: int = Field(..., title="Clicks", description="The number of clicks on the short URL")
    created_at: datetime = Field(..., title="Created At", description="The date and time when the short URL was created")

@app.on_event("startup")
async def startup():
    print("Starting up...")
    global db_pool
    db_pool = await create_pool(db_name="urls")


@app.get("/")
def health():
    return {"status": "ok"}

url_router = APIRouter()

@url_router.post("/")
async def create_short_url(url: URL):
    """
    Create a short URL
    """
    short_code = await get_next_id()
    short_code = encode(short_code)
    url.short_code = short_code
    url.expiration_date = datetime.now() + timedelta(days=365)
    async with db_pool.acquire() as conn:
        async with conn.cursor() as cursor:
            query = """INSERT INTO urls (short_code, long_url, expiration_date) VALUES (%s, %s, %s)"""
            await cursor.execute(query, (url.short_code, url.long_url, url.expiration_date))
            await conn.commit()
    return JSONResponse(content=url.json(), status_code=201)


@url_router.get("/{short_code}")
async def get_short_url(short_code: str):
    """
    Get the long URL from the short code
    """
    async with db_pool.acquire() as conn:
        async with conn.cursor() as cursor:
            query = """SELECT * FROM urls WHERE short_code = %s"""
            await cursor.execute(query, (short_code,))
            result = await cursor.fetchone()
            if result:
                url = URL(short_code=result[0], long_url=result[1], expiration_date=result[2])
                return JSONResponse(content=url.dict(), status_code=200)
            else:
                raise HTTPException(status_code=404, detail="Short URL not found")

@url_router.get("/")
async def get_short_urls():
    """
    Get all short URLs
    """
    async with db_pool.acquire() as conn:
        async with conn.cursor() as cursor:
            query = """SELECT * FROM urls"""
            await cursor.execute(query)
            result = await cursor.fetchall()
            urls = [URL(short_code=row[0], long_url=row[1], expiration_date=row[2]) for row in result]
            return JSONResponse(content=[url.dict() for url in urls], status_code=200)


@url_router.delete("/{short_code}")
async def delete_short_url(short_code: str):
    """
    Delete a short URL
    """
    async with db_pool.acquire() as conn:
        async with conn.cursor() as cursor:
            query = """DELETE FROM urls WHERE short_code = %s"""
            await cursor.execute(query, (short_code,))
            await conn.commit()
            if cursor.rowcount == 0:
                raise HTTPException(status_code=404, detail="Short URL not found")
    return JSONResponse(content={"detail": "Short URL deleted"}, status_code=200)

@url_router.get("/analytics/{short_code}")
async def get_url_analytics(short_code: str):
    """
    Get analytics for a short URL
    """
    pass

@url_router.put("/{short_code}")
async def update_short_url(short_code: str, url: URL):
    """
    Update a short URL
    """
    async with db_pool.acquire() as conn:
        async with conn.cursor() as cursor:
            query = """UPDATE urls SET long_url = %s, expiration_date = %s WHERE short_code = %s"""
            await cursor.execute(query, (url.long_url, url.expiration_date, short_code))
            await conn.commit()
            if cursor.rowcount == 0:
                raise HTTPException(status_code=404, detail="Short URL not found")
    return JSONResponse(content=url.dict(), status_code=200)



app.include_router(router=url_router, prefix="/api/v1/urls", tags=["URL Shortener"])


if __name__ == "__main__":
    uvicorn.run("main:app", host="0.0.0.0", port=8000, reload=True)
