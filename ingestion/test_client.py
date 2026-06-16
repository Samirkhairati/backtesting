import grpc
import ingestion_pb2
import ingestion_pb2_grpc
import sys
import uuid

def run():
    # Use the docker internal networking to hit the running service
    print("Connecting to ingestion service at localhost:50052...")
    with grpc.insecure_channel('localhost:50052') as channel:
        stub = ingestion_pb2_grpc.IngestionServiceStub(channel)
        
        # Generate a random experiment ID so we can track it
        exp_id = str(uuid.uuid4())
        
        request = ingestion_pb2.IngestDataRequest(
            experiment_id=exp_id,
            symbols=["AAPL", "MSFT"],
            start_date="2024-01-01",
            end_date="2024-01-10"
        )
        
        print(f"Sending IngestDataRequest for experiment {exp_id}...")
        try:
            response = stub.IngestData(request)
            print(f"Response success: {response.success}")
            print("Check the docker logs of the ingestion container to see it working in the background!")
        except Exception as e:
            print(f"Error: {e}")

if __name__ == '__main__':
    run()
