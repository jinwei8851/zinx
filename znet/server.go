package znet

import (
	"fmt"
	"gocode/zinx/ziface"
	"net"
)

//定义一个server的服务器模块，实现Isserver的接口
type Server struct {
	//服务器名称
	Name string
	//服务器绑定的ip版本
	IPVersion string
	//监听的ip
	IP string
	//监听的端口多少
	Port int
	//当前的server添加一个router，server注册的链接对应的处理业务
	Router ziface.IRouter
}

//定义当前客户端链接的绑定的handle api，以后优化由用户自定义handle方法
//func CallBackClient(conn *net.TCPConn, data []byte, cnt int) error {
//	// 回显业务
//	fmt.Println("[Conn Handle] callbackToclient...")
//	if _, err := conn.Write(data[:cnt]); err != nil {
//		fmt.Println("write back buf err", err)
//		return errors.New("CallBackToclient error")
//	}
//	return nil
//}

func (s *Server) Start() {
	fmt.Printf("[START] Server listenner at IP: %s, Port %d, is starting\n", s.IP, s.Port)
	go func() {
		//1 获取一个TCP的Addr
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Println("resolve tcp addr error:", err)
			return
		}
		//2 监听服务器的地址
		listener, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			fmt.Println("listen", s.IPVersion, "err", err)
			return
		}
		fmt.Println("start Zinx server succ", s.Name, "succ,Listenning")
		var cid uint32
		cid = 0

		//3 阻塞的等待客户端进行连接，处理客户端链接读写业务
		for {
			conn, err := listener.AcceptTCP()
			if err != nil {
				fmt.Println("Accept err", err)
				continue
			}
			//处理新链接的业务方法和conn进行绑定，得到我们的链接模块
			dealConn := NewConnection(conn, cid, s.Router)
			cid++

			//启动当前业务
			go dealConn.Start()
		}
	}()
}
func (s *Server) Stop() {
	//TOOD 将服务器的资源、状态或者一些已经开辟的链接信息，进行停止或者回收

}
func (s *Server) Server() {
	//启动server服务的功能
	s.Start()

	//TOOD做一些启动服务器之后的额外业务

	//阻塞状态
	select {}
}

/*
 初始化Server模块
*/
func NewServer(name string) ziface.IServer {
	s := &Server{
		Name:      name,
		IPVersion: "tcp4",
		IP:        "0.0.0.0",
		Port:      8999,
		Router:    nil,
	}
	return s
}

//路由功能：给当前服务注册一个路由业务方法，供客户端链接处理使用
func (s *Server) AddRouter(router ziface.IRouter) {
	s.Router = router

	fmt.Println("Add Router succ! ")
}
