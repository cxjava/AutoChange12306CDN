# AutoChange12306CDN

一个自动切换12306 CDN的代理，只需设置浏览器的代理为此软件监听端口，每次查询请求都会更换CDN，达到快速刷票的目的。

思路来自[分享12306秒票杀手锏源码](http://www.cnblogs.com/guozili/p/3512490.html)

# 推荐使用此方法 [Fiddler设置教程](https://git.oschina.net/charles/AutoChange12306CDN/wikis/fiddler)

# 使用方法

* 查找自己当前环境最快的CDN地址，运行findIP/findIP.exe，相同目录下面会生成12306_ip.txt，IP是Ping速度由快到慢进行排列，只需选取前30个IP写入到config.ini里面的cdn= []
* 添加自己需要查票的起点站和终点站到config.ini，请仔细阅读config.ini里面的相关项配置，如果不清楚，不要随意修改配置
* 打开本软件，设置谷歌浏览器的代理地址为本软件监听地址：127.0.0.1:8080，修改代理的软件有[Proxy SwitchySharp](https://chrome.google.com/webstore/detail/dpplabbmogkhghncfbfdeeokoefdjegm) 安装教程请参考 [谷歌 Chrome 配合 SwitchySharp 扩展](https://github.com/goagent/goagent/blob/wiki/InstallGuide.md#%E6%B5%8F%E8%A7%88%E5%99%A8%E8%AE%BE%E7%BD%AE%E6%96%B9%E6%B3%95)
* 如果使用谷歌浏览器订票，请在订票页面，按F12，选择下面的console栏，在光标位置输入：window.autoSearchTime = 2000; 再按回车。此操作为了设置每次查询间隔为2秒，官网默认为5秒，自己可以设置其他值。
* 如果使用[12306订票助手.NET版](http://www.fishlee.net/soft/12306/#C-308)订票，请设置其中的代理地址为本软件的监听地址:127.0.0.1:8080
* 谷歌浏览器如果通过本软件访问： https://kyfw.12306.cn/otn/leftTicket/init ，会出现证书错误问题，消息如下：

> #### 您的连接不是私密连接
> 
> #### 攻击者可能会试图从kyfw.12306.cn窃取您的信息（例如：密码、通讯内容或信用卡信息）。
> 
> #### 高级
点击 高级 ，再次点击 "继续kyfw.12306.cn（不安全）" 

# 打包好的下载地址
链接：[http://pan.baidu.com/s/1qW8tMhA](http://pan.baidu.com/s/1qW8tMhA) 密码: ua1x

#### 最后希望大家都能早日买到火车票票回家团圆
