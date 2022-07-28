package znet

import (
	"errors"
	"fmt"
	"gocode/zinx/utils"
	"gocode/zinx/ziface"
	"io"
	"net"
	"sync"
)

/*
  链接模块
*/
type Connection struct {
	//当前conn路属于哪个server
	TcpServer ziface.IServer
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
	//Router ziface.IRouter
	MsgHandler ziface.IMsgHandle
	//无缓冲管道，用于读、写两个goroutine之间的消息通信
	msgChan chan []byte

	//链接属性集合
	property map[string]interface{}
	//保护链接属性修改的锁
	propertyLock sync.RWMutex
}

// 初始化链接模块的方法
func NewConnection(server ziface.IServer, conn *net.TCPConn, connID uint32, msgHandler ziface.IMsgHandle) *Connection {
	c := &Connection{
		TcpServer: server,
		Conn:      conn,
		ConnID:    connID,
		//handleAPI: callback_api,
		//Router:   router,
		MsgHandler: msgHandler,
		isClosed:   false,
		msgChan:    make(chan []byte),
		ExitChan:   make(chan bool, 1),
		property:   make(map[string]interface{}), //对链接属性map初始化
	}
	//将conn加入到connmgr中
	c.TcpServer.GetConnMgr().Add(c)
	return c
}

//设置链接属性
func (c *Connection) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	c.property[key] = value
}

//获取链接属性
func (c *Connection) GetProperty(key string) (interface{}, error) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	if value, ok := c.property[key]; ok {
		return value, nil
	} else {
		return nil, errors.New("no property found")
	}
}

//移除链接属性
func (c *Connection) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	delete(c.property, key)
}

//链接的读业务方法
func (c *Connection) StartReader() {
	fmt.Println("[Reader Goroutine is running ...]")
	defer fmt.Println("connID = ", c.ConnID, "Reader is exit, remote addr is", c.RemoteAddr().String())
	defer c.Stop()

	for {
		//读取客户端的数据到buf中，最大512字节
		//buf := make([]byte, utils.GlobalObject.MaxPacketSize)
		//_, err := c.Conn.Read(buf)
		//if err != nil {
		//	fmt.Println("recv buf err", err)
		//	continue
		//}

		// 创建拆包解包的对象
		dp := NewDataPack()

		//读取客户端的Msg head 二进制流八个字节
		headData := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(c.GetTCPConnection(), headData); err != nil {
			fmt.Println("read msg head error ", err)
			break
			//c.ExitBuffChan <- true
			//continue
		}

		//拆包，得到msgid 和 datalen 放在msg中
		msg, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("unpack error ", err)
			break
			//c.ExitBuffChan <- true
			//continue
		}
		//根据dataLen 再次读取data，放在msg.data
		var data []byte
		if msg.GetMsgLen() > 0 {
			data = make([]byte, msg.GetMsgLen())
			if _, err := io.ReadFull(c.GetTCPConnection(), data); err != nil {
				fmt.Println("read msg data error ", err)
				break
				//c.ExitBuffChan <- true
				//continue
			}
		}
		msg.SetData(data)

		//得到当前conn数据的request数据
		req := Request{
			c,
			msg,
		}
		if utils.GlobalObject.WorkerPoolSize > 0 {
			//已经启动工作池机制，将消息交给Worker处理
			c.MsgHandler.SendMsgToTaskQueue(&req) //交给worker
		} else {
			//从绑定好的消息和对应的处理方法中执行对应的Handle方法
			go c.MsgHandler.DoMsgHandler(&req)
		}
		//从路由Routers 中找到注册绑定Conn的对应Handle,执行注册的路由方法
		//go func(request ziface.IRequest) {
		//	//执行注册的路由方法
		//	c.Router.PreHandle(request)
		//	c.Router.Handle(request)
		//	c.Router.PostHandle(request)
		//}(&req)
		//go c.MsgHandler.DoMsgHandler(&req)

		//// 调用当前链接所绑定的HandleAPI
		//if err := c.handleAPI(c.Conn, buf, cnt); err != nil {
		//	fmt.Println("ConnID", c.ConnID, "handle is error", err)
		//	break
		//}
	}
}

//提供一个sendmsg方法 我们要发送给客户端的数据，先进行封包
/*
	写消息Goroutine，专门给客户端发送消息的模块
*/
func (c *Connection) StartWriter() {
	fmt.Println("[Writer Goroutine is running]")
	defer fmt.Println(c.RemoteAddr().String(), "[conn Writer exit!]")

	for {
		select {
		case data := <-c.msgChan:
			if _, err := c.Conn.Write(data); err != nil {
				fmt.Println("Send Data error:, ", err, " Conn Writer exit")
				return
			}
		case <-c.ExitChan:
			//conn已经关闭
			return
		}
	}

}

func (c *Connection) Start() {
	fmt.Println("Conn Start()... ConnID=", c.ConnID)
	//启动从当前链接的读业务
	//TODO 启动当前链接写数据的业务
	go c.StartReader()
	go c.StartWriter()

	//按照开发者传递进来，创建连接函数，执行相应hook函数
	//==================
	//按照用户传递进来的创建连接时需要处理的业务，执行钩子方法
	c.TcpServer.CallOnConnStart(c)
	//==================
}

//停止链接 结束当前链接的工作
func (c *Connection) Stop() {
	fmt.Println("Conn Stop()... ConnID=", c.ConnID)
	if c.isClosed == true {
		return
	}
	c.isClosed = true
	//==================
	//如果用户注册了该链接的关闭回调业务，那么在此刻应该显示调用
	c.TcpServer.CallOnConnStop(c)
	//==================

	// 关闭socker 链接
	c.Conn.Close()

	//告知writer关闭
	c.ExitChan <- true

	//当前连接从connmgr中删除掉
	c.TcpServer.GetConnMgr().Remove(c)
	//关闭管道
	close(c.ExitChan)
	close(c.msgChan)
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

//// 发送数据， 将数据发送给远程客户端
//func (c *Connection) Send(data []byte) error {
//	return nil
//}
func (c *Connection) SendMsg(msgId uint32, data []byte) error {
	if c.isClosed == true {
		return errors.New("Connection closed when send msg")
	}
	//将data封包，并且发送
	dp := NewDataPack()
	msg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id = ", msgId)
		return errors.New("Pack error msg ")
	}
	//写回客户端
	c.msgChan <- msg
	//if _, err := c.Conn.Write(msg); err != nil {
	//	fmt.Println("Write msg id ", msgId, " error ")
	//	//c.ExitBuffChan <- true
	//	return errors.New("conn Write error")
	//}

	return nil
}
