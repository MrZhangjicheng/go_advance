# reids-benchmark 

### 1. 使用 redis benchmark 工具, 测试 10 20 50 100 200 1k 5k 字节 value 大小，redis get set 性能

#### redis配置  4c12G 容器启动  内网环境测试 

#### 命令：
    ./redis-benchmark -h XXX -p XXXX -d 10 -t get,set 

##### SET:

| 字节数 | 执行次数和耗时|每秒请求次数|  
|:--:|:--:|:--:|
|10|100000 requests completed in 12.16 seconds|8224.36 requests per second|
|20|100000 requests completed in 11.22 seconds|8914.25 requests per second|
|50|100000 requests completed in 18.34 seconds|5453.45 requests per second|
|100|100000 requests completed in 12.94 seconds|7725.59 requests per second|
|200|100000 requests completed in 12.20 seconds|8193.36 requests per second|
|1k|100000 requests completed in 11.14 seconds|8980.69 requests per second|
|5k|100000 requests completed in 17.96 seconds| 5567.62 requests per second|


##### GET:

| 字节数 | 执行次数和耗时|每秒请求次数|  
|:--:|:--:|:--:|
|10|100000 requests completed in 10.62 seconds|9418.86 requests per second|
|20|100000 requests completed in 10.44 seconds|9578.54 requests per second|
|50|100000 requests completed in 12.51 seconds|7992.33 requests per second|
|100|100000 requests completed in 11.53 seconds|8671.52 requests per second|
|200|100000 requests completed in 11.31 seconds|8844.86 requests per second|
|1k|100000 requests completed in 11.49 seconds|8706.25 requests per second|
|5k| 100000 requests completed in 19.67 seconds|5083.11 requests per second|


