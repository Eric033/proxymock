![图片](https://github.com/Eric033/proxymock/blob/main/logo.png)
###
proxymock是一款支持带报文头长度的tcp xml协议挡板
* 支持网络超时异常模拟
* 支持mock常规功能，包括模板预埋，规则增删改查
* 支持规则定时清理
* 支持代理透传模式
* REST API
* 部署便捷
# 用户手册
[Proxymock5分钟入门](https://github.com/Eric033/proxymock/wiki/%E4%BD%BF%E7%94%A8%E8%AF%B4%E6%98%8E)
# 如何启动
* cp 控制命令监听端口，用来接收规则增删改查请求
* enc 指定报文编码格式
* exp 指定规则默认的清理时间
* mp 挡板端口
* pre tcp xml协议报文头字段长度（标识报文长度）
* test 测试模式，将在8808端口启动一个tcp echo 挡板，便于工具测试/学习使用，该
* 挡板将返回收到的报文
``` 
PS E:\dev\jjmock> .\main.exe -help
Usage of E:\dev\jjmock\main.exe:
  -cp string
        command port(http default 9091) (default "9091")
  -enc string
        mock encoding:utf-8 , gbk (default "utf-8")
  -exp int
        rule default expire time in second (default 3600)
  -mp string
        mock port(tcp default 9090) (default "9090")
  -pre int
        frame prefix length;-1 MCA (default 6)
  -test
        true : start with a echo sever for testing (default true)
```
# 原理
## 主流程
1. 收到请求
2. 依次匹配已有的规则，匹配成功将依次执行规则指定的动作
3. 返回应答
## 匹配优先级
* 优先匹配常规规则，当常规规则未命中将尝试执行默认规则（若存在）
* 当存在多个规则都满足的情况下，最早设置的规则生效。所以规则在使用完成后要及时删除，
规则清理时间要尽可能短，避免存在大量不再使用的规则。
# 规则
## 类型
   规则通过规则的编码字段是否为 default分为以下两种规则
* 常规规则
* 默认规则
## 组成
* 匹配条件 (必选，默认规则该字段为空)
  支持and or逻辑
  支持xpath 值匹配，xpath针对mock收到的请求报文
* 动作 (可选)
  1. 前置动作 pre
     用来修改请求报文
  2. 后置动作 post
     用来修改应答报文
  3. 挡板动作 mock
     用来返回指定模板内容
  4. 转发动作 forward
     将报文转发至指定地址
  5. 超时模拟 timeout
     sleep
* 清理时间 (可选)
  指定规则超时时间，超时将被删除
* 编号 (可选)
  指定规则id，除了指定默认规则外
# 报文模板
  供挡板动作使用，当保存模板的模板名称已存在将执行更新操作
## 组成
* 模板名称 templateName 
* 模板内容 data
