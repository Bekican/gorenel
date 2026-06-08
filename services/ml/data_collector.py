import json
import redis
import pandas as pd
import numpy as np
from datetime import datetime, timedelta
import logging

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class DataCollector:  # <-- BU SATIR ÇOK ÖNEMLİ
    def __init__(self, redis_host='localhost', redis_port=6379):
        self.redis_client = redis.Redis(host=redis_host, port=redis_port, decode_responses=True)
        self.stream_name = 'traffic_stream'
        self.consumer_group = 'ml_processors'
        self.consumer_name = 'data_collector'
        
        try:
            self.redis_client.xgroup_create(self.stream_name, self.consumer_group, id='0', mkstream=True)
            logger.info(f"Consumer group '{self.consumer_group}' oluşturuldu.")
        except redis.exceptions.ResponseError as e:
            if "BUSYGROUP" not in str(e):
                raise e
            logger.info(f"Consumer group '{self.consumer_group}' zaten mevcut.")

    def collect_batch(self, count=100):
        try:
            messages = self.redis_client.xreadgroup(
                self.consumer_group, self.consumer_name, {self.stream_name: '>'}, count=count
            )
            records = []
            if messages:
                for stream, msgs in messages:
                    for msg_id, data in msgs:
                        try:
                            record = json.loads(data.get('data', '{}'))
                            record['_id'] = msg_id
                            records.append(record)
                            self.redis_client.xack(self.stream_name, self.consumer_group, msg_id)
                        except json.JSONDecodeError:
                            continue
            return records
        except Exception as e:
            logger.error(f"Veri toplama hatası: {e}")
            return []

    def preprocess_data(self, records):
        if not records:
            return pd.DataFrame()
        df = pd.DataFrame(records)
        if 'timestamp' in df.columns:
            df['timestamp'] = pd.to_datetime(df['timestamp'])
            df['hour'] = df['timestamp'].dt.hour
        numeric_cols = ['response_time', 'request_size', 'response_size', 'status_code']
        for col in numeric_cols:
            if col in df.columns:
                df[col] = pd.to_numeric(df[col], errors='coerce')
        df.fillna(0, inplace=True)
        return df

if __name__ == '__main__':
    collector = DataCollector()
    print("Veri toplayıcı hazır...")