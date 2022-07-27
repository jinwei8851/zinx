package ziface

/*
	连接管理模块
*/
type IConnManager interface {
	///添加
	Add(conn IConnection)
	//删除
	Remove(conn IConnection)
	//根据connID获取连接
	Get(connID uint32) (IConnection, error)
	//获取当前连接
	Len() int
	//删除并停止所有链接
	ClearConn()
}
