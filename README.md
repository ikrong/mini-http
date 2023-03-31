## 一个非常轻量的静态资源服务器

这个静态服务器的主要目的就是为了部署静态资源，也只具有静态访问功能，不适用于接口开发或者资源重定向。

项目地址：

* [https://github.com/ikrong/mini-http](https://github.com/ikrong/mini-http)
* [https://hub.docker.com/r/ikrong/mini-http](https://hub.docker.com/r/ikrong/mini-http)

目前支持的功能如下

- [x] 支持静态资源访问
- [x] 单页面路由
- [x] 支持端口自定义
- [x] 支持https配置
- [x] 支持多域名配置

## 使用方法

```shell
docker run -ti --rm --init -p 80:80 ikrong/mini-http
```

## 示例

1. 部署一个本地静态服务器，你可以这样启动

```shell
docker run -ti --rm --init -p 80:80 ikrong/mini-http
```

2. 使用本地静态文件

```shell
docker run -ti --rm --init \
    -p 80:80 \
    -v /local/dir/:/www/ \
    ikrong/mini-http
```

3. 使用其他端口

```shell
docker run -ti --rm --init \
    -p 8080:80 \
    -v /local/dir/:/www/ \
    ikrong/mini-http
```

4. 启动 `https` 服务器

```shell
docker run -ti --rm --init \
    -p 443:443 \
    -v /localhost/cert/:/cert/ \
    ikrong/mini-http \
    /serve \
        --domain localhost \
        --cert /cert/cert.pem \
        --key /cert/cert.key
```

> --domain 参数是告诉程序通过 `localhost` 域名访问服务
> 
> --cert 参数是告诉程序使用哪个证书
> 
> --key 参数是告诉程序使用哪个证书私钥
> 

4. 同时绑定多个域名


```shell
docker run -ti --rm --init \
    -p 80:80 \
    ikrong/mini-http \
    /serve \
        --domain localhost \
        --domain example.com
```

> 可以指定多对 domain 参数来绑定多个域名

5. 多个域名指定多个静态资源

```shell
docker run -ti --rm --init \
    -p 80:80 \
    -v /res/:/html/ \
    ikrong/mini-http \
    /serve \
        --domain localhost \
        --root /html/localhost/:/html/localhost/ \
        --domain example.com \
        --root /html/example.com/:/html/example.com/
```

> 每一个 domain 后面都可以依次跟一个 root 来指定静态资源目录
>
> domain cert key root 这四个参数 domain 是必选，其他三个选填
>
> 可以设置多组，以支持多个域名多个静态资源

6. 单页面应用

```shell
docker run -ti --rm --init \
   -p 80:80 \
   ikrong/mini-http \
   /serve \
     --domain localhost \
     --mode history
```

> mode 参数设置为 history 可以让对应的 domain 支持单页面应用访问

## MIT LICENSE