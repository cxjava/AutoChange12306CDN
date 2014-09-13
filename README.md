# AutoChange12306CDN

一个自动切换12306 CDN的代理，只需设置浏览器的代理为此软件监听端口，每次查询请求都会更换CDN，达到快速刷票的目的。

思路来自[分享12306秒票杀手锏源码](http://www.cnblogs.com/guozili/p/3512490.html)

# 使用方法

* 查找自己当前环境最快的CDN地址，具体方法可以参考[利用GoAgent多开抢票攻略,12306抢票攻略](http://matychen.iteye.com/blog/1988528)里面的PingInfoView,添加到配置文件config.ini里面,找不到很多IP的直接找到12306订票助手.NET版本的IP文件：12306订票助手.NET\Profile\cache\plugin\servernode.json（需启动一次后才能生成）
* 添加自己需要查票的起点站和终点站到config.ini
* 打开本软件，设置谷歌浏览器的代理地址为本软件监听地址，修改代理的软件有[Proxy SwitchySharp](https://chrome.google.com/webstore/detail/dpplabbmogkhghncfbfdeeokoefdjegm)
* 如果使用谷歌浏览器订票，请打开谷歌浏览器，按F12，选择下面的console栏，在光标位置输入：window.autoSearchTime = 2000; 再按回车。设置每次查询间隔为2秒，自己可以设置其他值。
* 如果使用12306订票助手.NET版订票，请设置代理为本软件的监听地址，默认为:127.0.0.1:8080

# 打包好的下载地址
链接：[http://pan.baidu.com/s/1qW8tMhA](http://pan.baidu.com/s/1qW8tMhA) 密码: ua1x

#### 最后希望大家都能早日买到火车票票回家团圆
