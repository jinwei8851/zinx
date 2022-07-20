package znet

import (
	"fmt"
	"gocode/zinx/ziface"
	"net"
)

/*
  链接模块
*/
type Connection struct {
	// 当前链接的socket TCP套接字
	Conn *net.TCPConn

	//链接ID
	ConnID uint32

	// 当前的链接状态
	isClosed bool

	//// 当前链接所绑定的处理业务方法API
	//handleAPI ziface.HandleFunc

	// 告知当前链接退出的停止的channel
	ExitChan chan bool

	// 该链接处理的方法router
	Router ziface.IRouter
}

// 初始化链接模块的方法
func NewConnection(conn *net.TCPConn, connID uint32, router ziface.IRouter) *Connection {
	c := &Connection{
		Conn:   conn,
		ConnID: connID,
		//handleAPI: callback_api,
		Router:   router,
		isClosed: false,
		ExitChan: make(chan bool, 1),
	}
	return c
}

//链接的读业务方法
func (c *Connection) StartReader() {
	fmt.Println("Reader Goroutine is running ...")
	defer fmt.Println("connID = ", c.ConnID, "Reader is exit, remote addr is", c.RemoteAddr().String())
	defer c.Stop()

	for {
		//读取客户端的数据到buf中，最大512字节
		buf := make([]byte, 512)
		_, err := c.Conn.Read(buf)
		if err != nil {
			fmt.Println("recv buf err", err)
			continue
		}

		//得到当前conn数据的request数据
		req := Request{
			c,
			buf,
		}

		//从路由Routers 中找到注册绑定Conn的对应Handle,执行注册的路由方法
		go func(request ziface.IRequest) {
			//执行注册的路由方法
			c.Router.PreHandle(request)
			c.Router.Handle(request)
			c.Router.PostHandle(request)
		}(&req)

		//// 调用当前链接所绑定的HandleAPI
		//if err := c.handleAPI(c.Conn, buf, cnt); err != nil {
		//	fmt.Println("ConnID", c.ConnID, "handle is error", err)
		//	break
		//}
	}
}

func (c *Connection) Start() {
	fmt.Println("Conn Start()... ConnID=", c.ConnID)
	//启动从当前链接的读业务
	//TODO 启动当前链接写数据的业务
	go c.StartReader()

}

//停止链接 结束当前链接的工作
func (c *Connection) Stop() {
	fmt.Println("Conn Stop()... ConnID=", c.ConnID)
	if c.isClosed == true {
		return
	}
	c.isClosed = true

	// 关闭socker 链接
	c.Conn.Close()

	//关闭管道
	close(c.ExitChan)
}

// 获取当前链接绑定的socket conn
func (c *Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}

// 获取当前链接模块的链接ID
func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}

// 获取远程客户端的TCP状态
func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

// 发送数据， 将数据发送给远程客户端
func (c *Connection) Send(data []byte) error {
	return nil
}
