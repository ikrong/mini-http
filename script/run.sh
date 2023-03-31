dir=`pwd`

echo "$dir/assets/:/www/"

docker run -ti --rm --init \
    --platform linux/amd64 \
    -v $dir/assets/:/www/ \
    -p 80:80 \
    -p 443:443 \
    ikrong/mini-http:latest \
    /serve \
        --port 80 \
        --https-port 443 \
        --domain example.com \
        --cert /www/cert/example.com/cert.pem \
        --key /www/cert/example.com/private.key \
        --root /www/cert/example.com/ \
        --domain example.net \
        --root /www/cert/example.net/ \
        --domain example.io \
        --cert /www/cert/example.io/cert.pem \
        --key /www/cert/example.io/private.key \
        --root /www/cert/example.io/ \
        --domain localhost \
        --mode history \
        --cert /www/cert/localhost/cert.pem \
        --key /www/cert/localhost/private.key \
        --root /www/cert/localhost/
