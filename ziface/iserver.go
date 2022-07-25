package ziface

//定义服务器接口
type IServer interface {
	//启动
	Start()
	//停止
	Stop()
	//运行
	Server()

	//路由功能：给当前的服务注册一个路由方法，供客户端的链接处理使用
	AddRouter(msgID uint32, router IRouter)
}
