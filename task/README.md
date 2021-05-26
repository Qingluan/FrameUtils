
## 任务系统 API Doc


### 部署规范

> Get
>> /task/v1 

> Post 
>> /task/v1/api
>> /task/v1/log

#### Push 部署任务

##### 单行string 格式

> example
    
    http:   
        http, https://www.baidu.com
        https://www.baidu.com
        http, www.baidu.com
    cmd:
        cmd , ls -lha /tmp && sleep 8 && ls /tmp
    
      
##### json 格式

 1. 带有encoding字段的
    多为多个任务打包

```json
{
    "input":"base64Str",
    "oper":"push",
    "tp":"http/cmd/....", # 规定了任务数据的类型 目前只支持 这三种
    "encode":"base64",
}
```



 2. 不带有encoding 字段的


```json
{
    "input":"....args ....",
    "oper":"push",
    "tp":"http/cmd....", # 规定了任务数据的类型 目前只支持 这三种
}
```

#### config 配置服务

```json
{
    "others":"xxx,xxxx,xxx,xxx",
    "proxy": "xxxxx",
    "taskNum" : "xxx",
    "try" : "xxx"
}
```

#### Pull 拉回任务信息

```json
{
    "id":"任务返回的id",
}

```

#### clear 清理node 的任务

```json

{
    "id":"任务返回的id 如为空则清理所有",
}

```


### API细节


#### 配置

> UpdateMyConfig

```go
allserver := ""
	if v, ok := try2str(data["others"]); ok {
		log.Println("Found Other:", utils.Green(v))
		ifsync = true
		allserver = v
	}

	if v, ok := try2str(data["proxy"]); ok {
		config.Proxy = v
	}
	if v, ok := try2int(data["try"]); ok {
		config.ReTry = v
	}
	if v, ok := try2int(data["taskNum"]); ok {
		config.TaskNum = v
	}
	if ifsync {
		info = config.SyncAllConfig(allserver, data)
	}
```

#### 从 字符串 部署

> DepatchByLines —— DepatchTask —— depatchTask


    line 分为两部分 :      tp , input
    会将tp 和input 分别组装为 TData （dict）：
    
               |

```json
    {
        "tp":tp,
        "input":input,
        "oper":"push"
    }
```
    从config.Others中随机选择一个node 作为api 转发json


#### 从 Tdata 部署

```json
    {
        "tp":tp,
        "input":input,
        "oper":"push"
    }
``` 



