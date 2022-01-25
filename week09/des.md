## socket 粘包解决方式

#### fix length:
    发送方和接受方进行约定,发送和接受固定长度的数据,并且该长度不超过缓冲区

#### delimiter based：
    发送方在每一次数据包结束时添加特殊的分隔符,来标识数据包的边界

#### length field based frame decoder：
    发送方,在数据包头部添加包的长度信息