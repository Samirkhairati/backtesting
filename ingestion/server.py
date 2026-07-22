import sys
import os
import time
import json
import logging
import uuid
import threading
from kafka import KafkaConsumer, KafkaProducer
from processor import JobProcessor

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

DB_URL = os.environ.get("DATABASE_URL")
if not DB_URL:
    raise ValueError("DATABASE_URL environment variable is not set and is required")

KAFKA_BROKERS = os.environ.get("KAFKA_BROKERS", "kafka:9092").split(",")
WORKER_ID = str(uuid.uuid4())

def heartbeat_loop(producer):
    while True:
        try:
            heartbeat = {
                "worker_id": WORKER_ID,
                "status": "ACTIVE",
                "timestamp": time.strftime('%Y-%m-%dT%H:%M:%SZ', time.gmtime())
            }
            producer.send('worker-heartbeats', heartbeat)
            producer.flush()
        except Exception as e:
            logger.error(f"Failed to send heartbeat: {e}")
        time.sleep(5)

if __name__ == '__main__':
    logger.info(f"Starting ingestion Kafka consumer... Worker ID: {WORKER_ID}")
    
    processor = JobProcessor(DB_URL)
    
    # Wait for Kafka to be ready
    producer = None
    for _ in range(10):
        try:
            producer = KafkaProducer(
                bootstrap_servers=KAFKA_BROKERS,
                value_serializer=lambda v: json.dumps(v).encode('utf-8'),
                key_serializer=lambda k: k.encode('utf-8') if k else None
            )
            break
        except Exception as e:
            logger.warning(f"Waiting for Kafka: {e}")
            time.sleep(3)
            
    if not producer:
        logger.error("Could not connect to Kafka producer")
        sys.exit(1)

    # Allow time for topics to be created by kafka-init container
    time.sleep(5)

    # Start heartbeat thread
    hb_thread = threading.Thread(target=heartbeat_loop, args=(producer,), daemon=True)
    hb_thread.start()

    consumer = KafkaConsumer(
        'ingestion-jobs',
        bootstrap_servers=KAFKA_BROKERS,
        group_id='ingestion-worker-group',
        value_deserializer=lambda m: json.loads(m.decode('utf-8')),
        enable_auto_commit=False,
        auto_offset_reset='earliest'
    )
    
    for message in consumer:
        job = message.value
        job_id = job.get('job_id')
        exp_id = job.get('experiment_id')
        symbol = job.get('symbol')
        start = job.get('start_date')
        end = job.get('end_date')
        
        logger.info(f"Processing job {job_id} for {symbol}")
        
        # Announce job started to claim it
        started_msg = {
            "job_id": job_id,
            "experiment_id": exp_id,
            "worker_id": WORKER_ID
        }
        producer.send('ingestion-job-events', value=started_msg, key=job_id)
        producer.flush()
        
        status = processor.process_job(symbol, start, end)
        
        producer.send('ingestion-completed-jobs', value={
            'job_id': job_id,
            'experiment_id': exp_id,
            'symbol': symbol,
            'status': status
        }, key=job_id)
        producer.flush()
        
        # Commit only after successful processing and publishing
        consumer.commit()

