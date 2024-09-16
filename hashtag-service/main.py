# Get API TO GET HASHTAG DATA

from fastapi import FastAPI
from fastapi.encoders import jsonable_encoder


# MOngoDB Connection
from pymongo import MongoClient

# Connection to MongoDB
client = MongoClient("mongodb://localhost:27017/")
db = client["instaphoto"]
collection = db["top_posts"]

app = FastAPI()


def get_hashtag_data(hashtag):
    data = collection.find_one({"hashtag": hashtag})
    if data:
        data["_id"] = str(data["_id"])
        return jsonable_encoder(data)
    return data


@app.get("/tag/")
async def get_hashtag(hashtag: str):
    return {
        "hashtag": hashtag,
        "data": get_hashtag_data(hashtag),
    }


@app.get("/hashtags")
async def get_hashtags():
    data = []
    for item in collection.find():
        item["_id"] = str(item["_id"])
        data.append(item)
    return data


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="0.0.0.0", port=8000)
