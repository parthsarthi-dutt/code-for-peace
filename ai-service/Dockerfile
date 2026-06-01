FROM python:3.10-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

# Generate the gRPC Python files if they are not pre-generated
RUN python -m grpc_tools.protoc -I./proto --python_out=./proto --grpc_python_out=./proto ./proto/evaluation.proto

EXPOSE 50051

CMD ["python", "server.py"]
