from fastapi import FastAPI, Request
from fastapi.responses import HTMLResponse, RedirectResponse, StreamingResponse
from fastapi.templating import Jinja2Templates

import logging
import os, time, uuid
from datetime import datetime
from threading import Thread
from faker import Faker

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)
fake = Faker()

app = FastAPI()

templates = Jinja2Templates(directory=os.path.join(os.path.dirname(__file__), "templates"))

os.system(f"rm -rf {os.path.join(os.path.dirname(__file__), 'data')}")
os.makedirs(os.path.join(os.path.dirname(__file__), "data"), exist_ok=True)

DATASETS_LOGS = os.path.join(os.path.dirname(__file__), "data")


def mock_deployment(deployment_id: str):
    filepath = os.path.join(DATASETS_LOGS, deployment_id + ".log")
    logger.info(f"initializing deployment {deployment_id}")
    logger.info(f"pushing logs to {filepath}")
    with open(filepath, "a", encoding="utf-8") as fp:
        for _ in range(100000):
            fp.write(f"{datetime.now().isoformat()}: INFO: {deployment_id} - {fake.text(max_nb_chars=64)}\n")
            fp.flush()
            time.sleep(0.5)


@app.get("/", response_class=HTMLResponse)
async def index_handler(request: Request):
    deployments = [x.split(".")[0] for x in os.listdir(DATASETS_LOGS) if x.endswith(".log")]
    return templates.TemplateResponse("index.html", {"request": request, "deployments": deployments})


@app.get("/deployments/{deployment_id}", response_class=HTMLResponse)
async def deployment_handler(request: Request, deployment_id: str):
    return templates.TemplateResponse("deployment.html", {"request": request, "deployment_id": deployment_id})


@app.post("/deployments")
async def create_deployment():
    deployment_id = uuid.uuid4().hex
    thread = Thread(target=mock_deployment, args=(deployment_id,))
    thread.start()

    return RedirectResponse(url=f"/deployments/{deployment_id}", status_code=301)


def log_tailer(deployment_id: str):
    filepath = os.path.join(DATASETS_LOGS, deployment_id + ".log")
    with open(filepath, "r", encoding="utf-8") as fp:
        while True:
            line = fp.readline()
            if not line:
                time.sleep(0.1)
                continue
            yield f"data: {line.strip()}\n\n"


@app.get("/logs/{deployment_id}")
async def logs_handler(deployment_id: str):
    return StreamingResponse(log_tailer(deployment_id), media_type="text/event-stream")
