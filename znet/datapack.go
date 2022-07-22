package znet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"gocode/zinx/utils"
	"gocode/zinx/ziface"
)

//封包，拆包具体模块
type DataPack struct {
}

//封包拆包实例化方法
func NewDataPack() *DataPack {
	return &DataPack{}
}

//获取包头长度方法
func (dp *DataPack) GetHeadLen() uint32 {
	// uint32(4字节) +  DataLen uint32(4字节)
	return 8
}

//封包方法
func (dp *DataPack) Pack(msg ziface.IMessage) ([]byte, error) {
	//创建一个存放bytes字节的缓冲
	dataBuff := bytes.NewBuffer([]byte{}) //从一个[]byte切片，构造一个Buffer

	//将datalen写进databuff中，littleendian 大端写大端读（内存的序列化地址）
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetMsgLen()); err != nil {
		return nil, err
	}
	//将msgid写进buf中
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetMsgId()); err != nil {
		return nil, err
	}

	//将data写进buf中
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetData()); err != nil {
		return nil, err
	}
	return dataBuff.Bytes(), nil
}

//数据流拆包方法
func (dp *DataPack) Unpack(binaryData []byte) (ziface.IMessage, error) {
	//创建一个从输入二进制数据的ioReader
	dataBuff := bytes.NewReader(binaryData)

	//只解压head的信息，得到dataLen和msgID
	msg := &Message{}

	//读datalen
	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.DataLen); err != nil {
		return nil, err
	}
	//读msgid
	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.Id); err != nil {
		return nil, err
	}
	//判断dataLen的长度是否超出我们允许的最大包长度
	if utils.GlobalObject.MaxPacketSize > 0 && msg.DataLen > utils.GlobalObject.MaxPacketSize {
		return nil, errors.New("Too large msg data recieved")
	}
	//这里只需要把head的数据拆包出来就可以了，然后再通过head的长度，再从conn读取一次数据
	return msg, nil
}
