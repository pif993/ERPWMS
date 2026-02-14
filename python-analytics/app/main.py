import os

from fastapi import FastAPI, Header, HTTPException
from pydantic import BaseModel, Field

app = FastAPI(title="analytics")

TOKEN = os.getenv("ANALYTICS_SERVICE_TOKEN", "")


def require_token(x_service_token: str) -> None:
    if not TOKEN or TOKEN in ("replace-token", "changeme", "change-me"):
        # hard fail if misconfigured
        raise HTTPException(status_code=500, detail="analytics token not configured")
    if x_service_token != TOKEN:
        raise HTTPException(status_code=401, detail="unauthorized")


class SlottingInput(BaseModel):
    warehouse_id: str = Field(min_length=1)
    max_moves: int = Field(default=100, ge=1, le=10000)


@app.get("/health")
async def health():
    return {"status": "ok"}


@app.get("/reports/stock-aging")
async def stock_aging(x_service_token: str = Header(default="")):
    require_token(x_service_token)
    return {"report": [], "read_only": True}


@app.post("/optimize/slotting")
async def optimize_slotting(payload: SlottingInput, x_service_token: str = Header(default="")):
    require_token(x_service_token)
    return {"status": "accepted", "input": payload.model_dump()}
