# http_moni
站点:https://moni.funwan.cn/      模拟POST/GET请求接口测试,文件上传文件下载

## 前端忘了哪里扒回来的，后端用Golang实现了POST和GET的模拟请求，帮助开发进行接口测试，程序不会记录任何接口数据。
## 支持HTTPS

- [演示站点](https://moni.funwan.cn/)


# 编译安装

```
go get zhaobisheng/http_moni
cd $GOPATH/github.com/zhaobisheng/http_moni
go build -o http_moni
```

# 启动方法

## 直接以运行二进制方式启动程序即可`./http_moni`
## 后台运行`nohuo ./http_moni &`
## 程序默认监听66端口，想修改程序监听的端口需要在main.go里面找到66，修改成你想要的端口即可
