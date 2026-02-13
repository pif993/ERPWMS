from fastapi import FastAPI, Header, HTTPException
from pydantic import BaseModel, Field
import os

app = FastAPI(title="analytics")
TOKEN = os.getenv("ANALYTICS_SERVICE_TOKEN", "replace-token")

class SlottingInput(BaseModel):
    warehouse_id: str = Field(min_length=1)
    max_moves: int = Field(default=100, ge=1, le=10000)

@app.get('/health')
async def health():
    return {"status": "ok"}

@app.get('/reports/stock-aging')
async def stock_aging(x_service_token: str = Header(default="")):
    if x_service_token != TOKEN:
        raise HTTPException(status_code=401, detail="unauthorized")
    return {"report": [], "read_only": True}

@app.post('/optimize/slotting')
async def optimize_slotting(payload: SlottingInput, x_service_token: str = Header(default="")):
    if x_service_token != TOKEN:
        raise HTTPException(status_code=401, detail="unauthorized")
    return {"status": "accepted", "input": payload.model_dump()}
