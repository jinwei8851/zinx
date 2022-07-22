package znet

import (
	"fmt"
	"io"
	"net"
	"testing"
)

// 开关utils/globalobj.go里面reload 关掉
func TestDataPack(t *testing.T) {
	/*
		模拟的服务器
	*/
	// 创建socket TCP
	listener, err := net.Listen("tcp", "127.0.0.1:7777")
	if err != nil {
		fmt.Println("server listen err", err)
		return
	}

	//创建一个go承载，负责从客户端处理业务
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println("server accept error", err)
			}
			go func(conn net.Conn) {
				dp := NewDataPack()
				for {
					//第一次从conn读，把包的head 读出来
					headData := make([]byte, dp.GetHeadLen())
					_, err := io.ReadFull(conn, headData)
					if err != nil {
						fmt.Println("read head error")
						return
					}
					msgHead, err := dp.Unpack(headData)
					if err != nil {
						fmt.Println("server umpake error", err)
						return
					}
					if msgHead.GetMsgLen() > 0 {
						//第二次head中的datalen的data
						msg := msgHead.(*Message)
						msg.Data = make([]byte, msg.GetMsgLen())

						//根据datalen的长度再次读取
						_, err := io.ReadFull(conn, msg.Data)
						if err != nil {
							fmt.Println("server unpake data error", err)
							return
						}
						fmt.Println("-->Recv MsgID:", msg.Id, "msgdatalen=", msg.DataLen, "data", string(msg.Data))

					}

				}

			}(conn)
		}
	}()
	//客户端读取数据，拆包处理

	/*
		模拟客户端
	*/
	conn, err := net.Dial("tcp", "127.0.0.1:7777")
	if err != nil {
		fmt.Println("client conn error", err)
		return
	}

	dp := NewDataPack()
	//模拟粘包过程
	msgl := &Message{
		Id:      1,
		DataLen: 4,
		Data:    []byte{'a', 'b', 'd', 'c'},
	}
	sendData1, err := dp.Pack(msgl)
	if err != nil {
		fmt.Println("client pack msgl error", err)
		return
	}
	msg2 := &Message{
		Id:      2,
		DataLen: 7,
		Data:    []byte{'z', 'i', 'n', 'x', 'i', 'n', 'x'},
	}
	sendData2, err := dp.Pack(msg2)
	if err != nil {
		fmt.Println("client pack msgl error", err)
		return
	}

	//两个包拼接在一起,...是切片打散
	sendData1 = append(sendData1, sendData2...)

	//一次性回写
	conn.Write(sendData1)

	//客户端阻塞
	select {}

}
