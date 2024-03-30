# HTTP 协议面试题

本文档中的内容主要摘自网络，内容可能不会十分严谨，请自行评估。

## 1. 简述 HTTP 协议

- HTTP 是一种无状态的、应用层的、基于请求-响应模式的超文本传输协议。
    - 无状态：每次请求都是独立的，服务器不会记录上一次请求的状态
    - **超文本**包含文本、图片、音频、视频等
- HTTP 协议使用 TCP 协议作为底层传输协议
- 一个HTTP请求包含：请求行、请求头、空行和消息体
    - 其中**请求行**包含：请求方法、请求URL、协议版本
- 一个HTTP响应包含：状态行、响应头、空行和消息体
    - 其中**状态行**包含：协议版本、状态码、状态码描述，例如: HTTP/1.1 200 OK

**请求方法**

英文叫Method，又叫HTTP动作，HTTP协议有7种请求方法：

- GET：请求服务器上的资源，不修改服务器数据
- POST：向服务器提交数据，创建或更新资源
- PUT：用于更新服务器上的现有资源或创建新资源
- HEAD：与GET方法类似，但服务器在响应HEAD请求时不应包含资源的主体内容，只返回头部信息
- DELETE：删除服务器资源
- OPTIONS：获取服务器支持的HTTP方法列表，常用于跨域资源共享（CORS）的预检请求
- TRACE：用于回显服务器收到的请求，服务器将原始请求报文作为响应返回给客户端，用于检查或调试请求的传输过程
    - 由于安全原因，大部分服务器禁用了TRACE方法
- CONNECT：主要用于代理服务器，用于建立到目标服务器的隧道。CONNECT方法允许客户端通过代理服务器创建一个到目标服务器的加密连接。
- PATCH：用于对资源进行部分更新

> - HTTP/1.0 定义了三种请求方法： GET, POST 和 HEAD方法
> - HTTP/1.1 新增了五种请求方法：OPTIONS, PUT, DELETE, TRACE 和 CONNECT 方法

### 1.2 HTTP 发展史

**HTTP/1/2 发展史**：

- 1991年 HTTP/0.9 发布 （只接受GET方法，不支持header）
- 1996年 HTTP/1.0 发布 （基本成型，支持富文本，header，状态码，缓存等）
- 1999年6月 HTTP/1.1 RFC2616 发布（使用了20多年的主流标准，支持连接复用，分块发送）
- 2009年Google发布 SPDY 协议（HTTP/2前身），后改进并作为HTTP/2标准
- 2014年12月 HTTP/2 标准提交（头部压缩HPACK、二进制分帧传输、服务器推送）
- 2015年2月 HTTP/2 标准被批准
- 2015年5月 HTTP/2 以RFC7540 发布

**HTTP/3 发展史**：

- 2013年google推出 QUIC（Quick UDP Internet Connections，目标替代TCP）
- 2018年10月提出将HTTP-OVER-QUIC更名为HTTP/3请求
- 2018年11月同意此请求 2022年6月HTTP/3以 RFC9114发布（UDP/QUIC/QPACK）

### 1.3 常用请求头

- Host：指定请求的目的地，即服务器的域名和端口号
    - 示例：`Host: example.com:8080`，若端口是80则可省略显示
- Accept：指定客户端可接受的内容类型，用于告诉服务器客户端期望接收的资源类型
    - 示例：`Accept: text/html`
- Accept-Encoding：指定客户端可接受的内容编码方式，用于告诉服务器客户端支持的压缩算法
    - 示例：`Accept-Encoding: gzip,deflate`
- Accept-Language：指定客户端偏好的语言列表
    - 示例：`Accept-Language: en-US,en;q=0.5`
- Authorization：用于进行身份验证，通常与用户名和密码一起发送，以允许客户端访问受保护的资源
    - 示例：`Authorization: Basic YWRtaW46YWRtaW4=`（基本认证示例）
- User-Agent：标识客户端的软件类型、操作系统、浏览器等信息，服务器可以根据此信息进行内容适配
    - 示例：`User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) Apple`
- Content-Type：指定请求或响应的实体主体的媒体类型（也称为 MIME 类型），用于告诉服务器请求或响应中包含的数据的格式
    - 示例：`Content-Type: application/x-www-form-urlencoded`
- Content-Length：指定请求或响应的实体主体的长度，以字节为单位
    - 示例：`Content-Length: 1337`
- Referer：指定从哪个URL链接过来的，用于告诉服务器请求的来源
    - 示例：`Referer: https://www.example.com/page.html`
- Cookie：包含了客户端的Cookie信息，用于在客户端和服务器之间传递状态信息
    - 示例：`Cookie: sessionId=abc123; username=johndoe`
- Cache-Control：指定缓存的行为，控制缓存的存储、过期等策略
    - 示例：`Cache-Control: no-cache`
- Connection：指定是否保持持久连接，用于告诉服务器是否要保持TCP连接打开以进行多个请求
    - 示例：`Connection: keep-alive`（希望在完成请求/响应后保持连接），或者值为`close`（希望~关闭连接）
- If-Modified-Since：用于条件GET请求，告诉服务器只有在指定日期之后被修改过的资源才会被发送，用于缓存优化
    - 示例：`If-Modified-Since: Sat, 29 Oct 1994 19:43:31 GMT`（当服务器资源最后修改时间未变化时返回304）
- If-None-Match：类似上个，用于缓存优化，包含资源的当前ETag值而不是最后修改时间，如果资源未更改，则服务器可以返回304状态码。
    - 示例：If-None-Match: "abcdefg"（当服务器资源Etag未变化时返回304）

### 1.4 常见的 Content-Type

- text/html：HTML格式的文本，通常用于网页内容。
- text/plain：纯文本格式，不包含任何格式化。
- application/json：JSON格式的数据，常用于Web服务和APIs。
- application/xml：XML格式的数据，用于传输XML文档。
- application/x-www-form-urlencoded：表单数据被编码为键值对，用于提交表单。
- multipart/form-data：用于表单提交时上传文件。
- image/png：PNG格式的图片。
- image/jpeg：JPEG格式的图片。
- audio/mpeg：MP3格式的音频文件。
- video/mp4：MP4格式的视频文件。
- application/octet-stream：未知类型的二进制数据，通常用于传输文件或其他类型的数据，浏览器收到此类型数据会提示用户下载该文件而不会渲染。

### 1.5 常用状态码

状态码表示服务器对请求的处理结果。常用的HTTP状态码有1xx、2xx、3xx、4xx、5xx。

#### 1.5.1 1xx系列

1xx是信息性状态码，用于传递一些信息，而不是表示请求成功或失败。

- 100 Continue：表示客户端可以继续发送请求的主体。这个状态码是用于客户端和服务器之间的交互，通常不会在浏览器中看到。
- 101 Switching Protocols：服务器同意切换协议。这通常用于升级协议，例如从HTTP切换到HTTPS。
- 102 Processing：表示服务器已经接收到了请求，但尚未完成处理。用于长时间处理的请求，例如分块传输编码的数据或WebDAV中的锁定操作。
- 103 Early Hints：此状态代码主要用于与 Link 链接头一起使用，以允许用户代理在服务器准备响应阶段时开始预加载 preloading 资源。

#### 1.5.2 2xx系列

2xx系列的状态码表示请求成功，并返回了响应。

- 200 OK：请求成功。服务器已成功处理请求并返回了请求的资源。
- 201 Created：请求成功并且服务器创建了新的资源。通常用于POST请求。
- 202 Accepted：服务器已接受请求，但尚未处理完成。
- 204 No Content：服务器成功处理了请求，但没有返回任何内容。例如，对于DELETE请求。

#### 1.5.3 3xx系列

3xx系列的状态码表示请求需要进一步处理才能完成。

- 301 Moved Permanently：请求的资源已被永久移动到新位置，并返回新的URL。
- 302 Found：请求的资源已被临时移动到新位置，并返回新的URL。
- 303 See Other：类似于302，但明确要求客户端使用GET方法获取资源。
- 304 Not Modified：表示请求的资源未修改，可以使用缓存的版本。

#### 1.5.4 4xx系列

4xx系列的状态码表示客户端发送的请求有错误。

- 400 Bad Request：请求格式错误，服务器无法理解。
- 401 Unauthorized：请求需要用户认证。
- 403 Forbidden：请求被禁止访问（服务器已知客户端身份，但仍禁止其访问）。
- 404 Not Found：请求的资源不存在。
- 405 Method Not Allowed：请求的方法不被允许。
- 413 Payload Too Large：请求的实体过大，服务器无法处理。
- 415 Unsupported Media Type：服务器不支持请求数据的媒体格式，因此服务器拒绝请求。
- 429 Too Many Requests：客户端发送的请求过多，服务器无法处理（限制请求速率）。
- 431 Request Header Fields Too Large：服务器不愿意处理请求，因为其头字段太大。在减小请求头字段的大小后，可以重新提交请求。

#### 1.5.5 5xx系列

5xx系列的状态码表示服务器在处理请求时发生错误。

- 500 Internal Server Error：服务器内部错误，无法处理请求。
- 501 Not Implemented：服务器不支持请求方法，因此无法处理
- 502 Bad Gateway：作为网关或者代理工作的服务器尝试执行请求时，从上游服务器收到无效的响应（上游错误）。
- 503 Service Unavailable：服务器暂时无法处理请求（由于超载或停机维护），一段时间后可能恢复正常。
- 504 Gateway Timeout：作为网关或者代理工作的服务器尝试执行请求时，未能及时从上游服务器收到响应。

### 1.6 缓存机制

客户端（通常是浏览器）可以将获取的资源保存在本地，下次请求相同的资源时可以直接从本地缓存中获取。
客户端缓存通过设置HTTP头部来控制缓存的行为，常用的头部包括：Cache-Control、Expires、ETag、Last-Modified等。

- Expires：这个头部字段提供了一个日期/时间，之后响应被认为过时。这是HTTP/1.0 的头部字段，在HTTP/1.1中已经被Cache-Control替代。
    - 如果 Cache-Control 头部字段存在，Expires 通常不会被使用。
- Cache-Control：这是最重要的缓存头部字段，它提供了关于如何缓存响应的指令。例如，`Cache-Control: max-age=3600`
  表示资源可以被缓存1小时。
    - 其支持的值包括：private、public、no-cache、max-age，no-store，默认为 private
    - private：客户端可以缓存响应，但只能在与原始服务器通信时使用该响应。
    - public：客户端和代理服务器都可以缓存响应。
    - no-cache：客户端可以缓存响应，但必须先与原始服务器验证其有效性。
        - 与 private/public 的区别在于它每次都要向原始服务器验证其有效性，而前者只在缓存过期时才验证。
    - no-store：客户端不能缓存响应，并且每次都要向原始服务器验证其有效性（**适用于极高隐私级别场景**）。
    - max-age：客户端可以缓存响应，但缓存的时间不能超过指定的秒数。
    - **如何验证缓存有效性？**
        - 使用 If-Modified-Since 或 If-None-Match 请求头来发出条件性请求
- Last-Modified：表示资源最后一次被修改的时间，通常与请求头 If-Modified-Since 配合使用
    - 客户端可以使用这个时间来发出条件性请求，即如果资源未更改，服务器可以返回 304 Not Modified。
- ETag：这是一个资源的特定版本的标识符，通常与请求头 If-None-Match 配合使用。
    - 客户端可以存储这个ETag值，并在后续的请求中使用 If-None-Match
      头部字段来验证资源是否已更改，服务端仅在Etag不匹配的情况下返回完整的资源，否则返回 304 Not Modified。
    - 若客户端同时设置了 If-Modified-Since 和 If-None-Match 头部字段，则优先使用后者，且当后者有变化时不再判断前者。

**强制缓存与协商缓存**

- 强制缓存：指的是Expires和Cache-Control，强制缓存表示在缓存有效时直接使用缓存数据，不请求服务器。
- 协商缓存：指的是Last-Modified和ETag，协商缓存表示在缓存失效时向服务器发出请求以验证缓存有效性，服务器验证缓存的确失效时返回新的资源，
  若缓存有效则返回304 Not Modified。

### 1.7 长连接

如果一个TCP连接仅完成一次HTTP请求和响应，则称为短连接。但这样会存在效率问题，若一个Web页面需要发出多个请求来下载资源，
这需要客户端与服务器建立多次TCP连接来完成任务，这样会导致整体过程十分耗时且消耗双端资源。

长连接可以解决这一问题，它允许在同一个TCP连接中传输多个HTTP请求和响应，从而减少了建立和关闭连接的消耗和延迟。
HTTP 长连接允许客户端通过请求头 `Connection: keep-alive` 告知服务器，希望本次请求完成后不要断开TCP连接，从而继续复用。

> HTTP/1.0 需要手动配置请求头来开启长连接，HTTP/1.1 默认开启长连接（不用设置请求头）。

**清理空闲的长连接**

若客户端建立了长连接，但后续没有新的请求，则服务器会等待一段时间后主动关闭连接。这通过服务器代理提供的 keepalive_timeout
参数来设置。

**Nginx设置HTTP长连接**

```shell
# nginx.conf

# 与客户端的长连接配置
http {
  keepalive_timeout 60s # default 75s;
  keepalive_requests 100 # default 100;
  keepalive_disable off;
}

location / {
  proxy_pass             http://your_upstream;
  proxy_read_timeout     300;
  proxy_connect_timeout  300;
  ...

  # 与上游服务器的长连接配置
  proxy_http_version 1.1;
  proxy_set_header Connection "";
}

# 与上游服务器的长连接配置
upstream backend {
    ip_hash;
    server backend1.example.com;
    server backend2.example.com;
    keepalive 32; # 配置每个worker进程中保留的空闲长连接的最大数量（不应过大，推荐为 QPS 的 10% ~ 20%）
    keepalive_timeout 60;
    keepalive_requests 200;
}
```

**设置不当导致Nginx出现大量TIME_WAIT**

两种情况：

- keepalive_requests 设置比较小，而 QPS 较大，导致Nginx频繁关闭与客户端的TCP连接
- keepalive 设置过小，而 QPS 较大，导致Nginx频繁关闭与上游的TCP连接

### 1.8 跨域请求

### 1.9 安全问题

## 2. 简述 HTTPS 协议

## 3. 简述 HTTP/2 协议

## 4. 简述 HTTP/3 协议
