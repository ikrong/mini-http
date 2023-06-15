dir=`pwd`

go run main.go \
   --port 80 \
   --https-port 443 \
   --domain example.com \
   --not-found $dir/assets/404.html \
   --cert $dir/assets/cert/example.com/cert.pem \
   --key $dir/assets/cert/example.com/private.key \
   --root $dir/assets/cert/example.com/ \
   --domain example.net \
   --root $dir/assets/cert/example.net/ \
   --domain example.io \
   --cert $dir/assets/cert/example.io/cert.pem \
   --key $dir/assets/cert/example.io/private.key \
   --root $dir/assets/cert/example.io/ \
   --domain localhost \
   --mode history \
   --cert $dir/assets/cert/localhost/cert.pem \
   --key $dir/assets/cert/localhost/private.key \
   --root $dir/assets/cert/localhost/