# AutoChange12306CDN

一个自动切换12306 CDN的代理，只需设置浏览器的代理为此软件监听端口，每次查询请求都会更换CDN，达到快速刷票的目的。

# 使用方法

* 软件启动后会查找当前环境最快的`12306 CDN`地址，`iprange.conf`是所有的12306的CDN地址。 
* 打开本软件，设置谷歌浏览器的代理地址为本软件监听地址：`127.0.0.1:8888`，修改代理的软件有[Proxy SwitchySharp](https://chrome.google.com/webstore/detail/proxy-switchyomega/padekgcemlokbadohgkifijomclgjgif) 安装配置教程请参考 [Chrome SwitchySharp配置](https://www.switchyomega.com/settings/)
* 如果使用谷歌浏览器订票，请在订票页面，按`F12`，选择下面的`console`栏，在光标位置输入：`window.autoSearchTime = 1000`; 再按回车。此操作为了设置每次查询间隔为`1`秒，官网默认为`5`秒，自己可以设置其他值,但是最好不要小于`1`秒（即`1000`），设置太小容易被封IP。
* 谷歌浏览器如果通过本软件访问： [https://kyfw.12306.cn/otn/leftTicket/init?linktypeid=dc](https://kyfw.12306.cn/otn/leftTicket/init?linktypeid=dc) ，会出现证书错误问题，消息如下：

> #### 您的连接不是私密连接
> 
> #### 攻击者可能会试图从kyfw.12306.cn窃取您的信息（例如：密码、通讯内容或信用卡信息）。
> 
> #### 高级
点击 *高级* ，再次点击 "继续kyfw.12306.cn（不安全）" 

# 打包好的下载地址

链接：[AutoChange12306CDN](https://github.com/cxjava/AutoChange12306CDN/releases)

## 备用方法(设置稍微复杂一丢丢)[Fiddler抢票设置教程](https://github.com/cxjava/AutoChange12306CDN/wiki/Fiddler%E8%AE%BE%E7%BD%AE%E6%95%99%E7%A8%8B)

链接：[Fiddler抢票需要的软件](https://github.com/cxjava/AutoChange12306CDN/releases/tag/v1.0.1)

# 最后希望大家都能早日买到火车票票回家团圆

思路来自[分享12306秒票杀手锏源码](http://www.cnblogs.com/guozili/p/3512490.html)
