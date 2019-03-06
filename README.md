# IAT Client
IAT Client是IAT管理测试平台对应的执行器，IAT Client无状态，理论上支持无限水平扩展。

#部署
+ 将Docker目录中的脚本下载到本地。
+ 将Build好的“iat-client”程序，存放到相同的目录下。
+ 执行“build.sh”脚本，构建docker镜像。
+ 执行“start.sh”脚本，启动docker容器。start.sh脚本中的IAT_SERVER填写IAT平台地址，IAT_CLIENT_NAME填写当前Client名称。

#依赖
https://github.com/terrytian0/iat