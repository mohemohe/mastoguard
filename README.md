# mastoguard

ウザいインスタンスをブロックするためのリバプロ

## 環境変数

| 名前 | デフォルト値 | 必須 | 説明 |
| - | - | - | - |
| PROXY_TARGET | \- | Y | プロキシ先のURL |
| LISTEN_ADDR | :8080 | N | mastoguardがlistenするアドレス |
| DENY_UA | \- | N | 部分一致で弾くUAを`,`区切りで指定する（`*`を指定すると全部弾く） |

## ログ形式

```
time:2020-02-07T08:08:01.961624Z        level:Debug     event:mastoguard start
time:2020-02-07T08:08:01.961640Z        level:Debug     env 'PROXY_TARGET':https://mohemohe.dev
time:2020-02-07T08:08:01.961643Z        level:Debug     env 'LISTEN_ADDR'::8080
time:2020-02-07T08:08:01.961646Z        level:Debug     env 'DENY_UA':高輝度うんこ,低純度鼻くそ
time:2020-02-07T08:08:01.961657Z        level:Info      event:mastoguard ready
time:2020-02-07T08:08:03.574684Z        level:Info      result:ALLOW    method:GET      url:https://mohemohe.dev/       remote:127.0.0.1:58598  useragent:curl/7.68.0
time:2020-02-07T08:08:03.580740Z        level:Info      result:DENY     method:GET      url:https://mohemohe.dev/       remote:127.0.0.1:58602  useragent:高輝度うんこ
time:2020-02-07T08:08:03.586146Z        level:Info      result:DENY     method:GET      url:https://mohemohe.dev/       remote:127.0.0.1:58604  useragent:低純度鼻くそ
time:2020-02-07T08:08:03.721094Z        level:Info      result:ALLOW    method:GET      url:https://mohemohe.dev/       remote:127.0.0.1:58606  useragent:うんこ
```