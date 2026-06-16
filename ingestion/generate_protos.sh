#!/bin/bash
python -m grpc_tools.protoc -I./core/proto/ingestion --python_out=. --grpc_python_out=. ./core/proto/ingestion/ingestion.proto
