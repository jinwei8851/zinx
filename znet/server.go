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
}

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

		//3 阻塞的等待客户端进行连接，处理客户端链接读写业务
		for {
			conn, err := listener.AcceptTCP()
			if err != nil {
				fmt.Println("Accept err", err)
				continue
			}
			//已经与客户端连接连接，做一些业务，做一个最基本的512字节长度的回显业务

			go func() {
				for {
					buf := make([]byte, 512)
					cnt, err := conn.Read(buf)
					if err != nil {
						fmt.Println("recv buf err", err)
						continue
					}
					fmt.Printf("recv client buf %s,cnt %d\n", buf, cnt)
					//回显功能
					if _, err := conn.Write(buf[:cnt]); err != nil {
						fmt.Println("write back buf err", err)
						continue
					}
				}
			}()
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
	}
	return s
}
