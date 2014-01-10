# AutoChange12306CDN

一个自动切换12306 CDN的代理，只需设置浏览器的代理为此软件监听端口，每次查询请求都会更换CDN，达到快速刷票的目的。

思路来自[分享12306秒票杀手锏源码](http://www.cnblogs.com/guozili/p/3512490.html)

# 使用方法

* 查找自己当前环境最快的CDN地址，具体方法可以参考[利用GoAgent多开抢票攻略,12306抢票攻略](http://matychen.iteye.com/blog/1988528)里面的PingInfoView,添加到配置文件里面
* 添加自己需要查票的城市地址，修改MiaoPiao.user.js 里面的stationGroups(162行)，自行添加
* 打开谷歌浏览器扩展安装界面 chrome://extensions/，把MiaoPiao.user.js 拖入到里面，不知道怎么操作的可以参考[安装说明 - 鱼の后花园](http://www.fishlee.net/soft/44/how_to_install.html)
* 打开本软件，设置谷歌浏览器的代理地址为本软件监听地址

# 最后希望大家都能早日买到火车票票回家团圆
