


### 部署函数
	这里正式从 task.call中通过callTp选取函数来执行 目前有  Httpcall / cmdcall / 后续还支持tcpcall 等...

> 如果 Others 存在,而 kargs 里没有特别指明 Local=trueOrAnyThing 则使用远端分布式部署任务
#### 支持调用类型

 * http
 * cmd
 * config
 * tcp