# mian_go_lib 自用的go小轮子

### tool 小工具

| 子包         | 描述      | 包含                     |
|------------|---------|------------------------|
| cipher     | 密码学包    | aes、sha256             |
| cmd_server | tcp服务器包 | cmd协程工具                |
| menu       | 通用菜单包   | 配置cmd gui，支持高阶键盘监听     |
| multi      | 多线程安全相关 | sync.map的泛型封装，支持任意类型的线程安全ID锁|
| misc       | 杂项包     | 一堆小工具。包含语法糖和函数式、协程编程组件 |
| spider     | 爬虫包     | 百度新闻、大盘、彩票             |
| token      | 鉴权包     | jwt等                   |

### x series 稍大一些的工具

| 子包       | 描述                         | 包含                                             |
|----------|----------------------------|------------------------------------------------|
| spider   | 爬虫包                        | 百度新闻、大盘、彩票                                     |
| xpush    | 推送包                        | 邮件、pushdeer、钉钉 sdk                             |
| xlog     | 通用日志包                      | 可以和push结合使用，并自主替换命令行输出                         |
| xstorage | 线程安全、使用方便得，支持复杂配置的自落盘cache | 支持int、float、bool、string和对应的slice，包含一个简单的web外包装 |
| xnews    | topic based缓存              | 支持根据topic配置限流器，定时清除，单条自定义默认过期时间                |
| xres     | 转表工具                       | 支持toml配置元信息进行excel处理                           |

### scaffold 脚手架 代码生成器相关

| 文件     | 描述                          |
|--------|-----------------------------|
| go_err | 自动将所有的硬编码error提取到单独的文件，并处理包 |
|        |                             |
|        |                             |

### 使用方法

可以参考test中的集成测试使用
