# s3-http-proxy

Little proxy to access an private S3 bucket via HTTP.

## Usage
```
export S3PROXY_BUCKET="nameofmybucket"
export S3PROXY_REGION="us-central-1"
export S3PROXY_PORT="3000"
./proxy
```

## Usage with Docker
```
docker run -e S3PROXY_BUCKET=nameofmybucket -p 3000:3000 --rm -it codemonauts/s3-http-proxy
```
