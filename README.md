# 基于Golang协程实现流量统计分析系统

## 并发编程
进程、线程、协程

go 协程   

```
message := make(chan string)
go func(){
   message <- "hello world"
   // 标准输出无法输➗
   //fmt.Println("hello world") 
}()
fmt.Println(<-message)
```

## 打点服务
网站 -> js上报打点数据 -> 打点服务器（nginx）-> access日志 -> 分析（go） -> 可视化

#### 打点服务器
1. 使用nginx的 ngx_http_empty_gif_module模块     
2. js post请求打点数据，利用nginx的access.log记录请求日志    
3. 关闭nginx的gzip，返回的1px的图片，没必要压缩，gzip需要占资源     


#### 统计分析模块

逐行消费日志 --> channel --> 日志解析 --> channel --> UV统计、PV统计 --> channel -->  数据存储

