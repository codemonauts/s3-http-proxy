# s3-http-proxy

Little proxy to access an private S3 bucket via HTTP.


## Usecase
When your application stores it's assets in an S3 bucket and you use e.g.
CloudFront, you can improve performance by configuring the bucket as a origin
and point a custom behaviour like '/assets' to the bucket. This way, the assets
get directly served from the bucket without shoving the request through your
application stack. This also work perfectly for privat buckets because
CloudFront can use an OAI (Origin Access Identity) to get permissions.  When you
now can't (for whatever reason) use CloudFront and just have a good old
webserver/reverseproxy like e.g. nginx in front of your application but still
wan't to directly serve assets from the bucket, you are out of luck because
nginx can't deal with IAM credentials to access a private bucket (and you don't
want to enable public access on your bucket!).  Because we had this scenario for
a few customers, we wrote this tool which you can run behind a
webserver/reverseproxy and then configure an location block for '/assets' which
routes the request to this tool, and get nearly the same behaviour as in the
setup with CloudFront (obviously it's not a full blown CDN but you still get
'direct' file access to the bucket without going through your app stack).


## Minimal usage example
```
export S3PROXY_BUCKET="nameofmybucket"
./s3-http-proxy
```

## Usage with Docker
```
docker run -e S3PROXY_BUCKET=nameofmybucket -p 3000:3000 --rm -it codemonauts/s3-http-proxy
```

## Configuration
All configuration happens via environment variables. 

| Name              | Required | Default             | Description                                                  |
| ----------------- | :------: | ------------------- | ------------------------------------------------------------ |
| S3PROXY_BUCKET    |    x     | -                   | Name of the bucket                                           |
| S3PROXY_REGION    |          | "eu-central-1"      | Region of the bucket                                         |
| S3PROXY_PORT      |          | 3000                | Listening port of the application                            |
| S3PROXY_CACHE     |          | ""                  | Set this to a path if you wan't the files to be cached       |
| S3PROXY_SIZELIMIT |          | 104857600 ( ~100MB) | Only files smaller than this are cached. Set to 0 to disable |
| S3PROXY_LOGGING   |          | "WARN"              | Loglevel ("ERROR","WARN","INFO","DEBUG")                     |


## Caching
This proxy can localy cache all files from S3 to disk for better performance. To
enable caching just set *S3PROXY_CACHING* to a valid path (relative or absolut
works both). The tool will then only do a HeadRequest to the bucket, when it has
the file already in it's cache to see if the file is still up to date
(Comparison of LastModified timestamp). If the file has changed in the  bucket
after we downloaded it, it will freshly get downloaded from the Bucket and
replaced on disk before a response is send.

If you don't need this invalidation check for your files, you can also directly
point your webserver to the cache directory of the plugin, because the files get
saved to disk with the same folder structure as in S3 so they can directly be
read and delivered by a webserver.


With ❤ by [codemonauts](https://codemonauts.com)
