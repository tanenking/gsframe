package zcommon

import (
	"errors"
	"sync"

	"github.com/tanenking/gsframe/gsinf"
	"github.com/tanenking/gsframe/internal/logger"
)

// ConnManager 连接管理模块
type ConnManager struct {
	connections map[uint32]gsinf.IConnection
	connLock    sync.RWMutex
}

// NewConnManager 创建一个链接管理
func NewConnManager() *ConnManager {
	return &ConnManager{
		connections: make(map[uint32]gsinf.IConnection),
	}
}

// Add 添加链接
func (connMgr *ConnManager) Add(conn gsinf.IConnection) {

	connMgr.connLock.Lock()
	//将conn连接添加到ConnMananger中
	connMgr.connections[conn.GetConnID()] = conn
	connMgr.connLock.Unlock()

	logger.Log().Debug("connection add to ConnManager successfully: conn num = %d", connMgr.Len())
}

// Remove 删除连接
func (connMgr *ConnManager) Remove(conn gsinf.IConnection) {

	connMgr.connLock.Lock()
	//删除连接信息
	delete(connMgr.connections, conn.GetConnID())
	connMgr.connLock.Unlock()
	logger.Log().Debug("connection Remove ConnID=%d, successfully: conn num = %d", conn.GetConnID(), connMgr.Len())
}

// Get 利用ConnID获取链接
func (connMgr *ConnManager) Get(connID uint32) (gsinf.IConnection, error) {
	connMgr.connLock.RLock()
	defer connMgr.connLock.RUnlock()

	if conn, ok := connMgr.connections[connID]; ok {
		return conn, nil
	}

	return nil, errors.New("connection not found")

}

// Len 获取当前连接
func (connMgr *ConnManager) Len() int {
	connMgr.connLock.RLock()
	length := len(connMgr.connections)
	connMgr.connLock.RUnlock()
	return length
}

// ClearConn 清除并停止所有连接
func (connMgr *ConnManager) ClearConn() {
	connMgr.connLock.Lock()

	//停止并删除全部的连接信息
	for connID, conn := range connMgr.connections {
		//停止
		conn.Stop()
		//删除
		delete(connMgr.connections, connID)
	}
	connMgr.connLock.Unlock()
	logger.Log().Debug("Clear All Connections successfully: conn num = %d", connMgr.Len())
}

// ClearOneConn  利用ConnID获取一个链接 并且删除
func (connMgr *ConnManager) ClearOneConn(connID uint32) {
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()

	connections := connMgr.connections
	if conn, ok := connections[connID]; ok {
		//停止
		conn.Stop()
		//删除
		delete(connections, connID)
		logger.Log().Debug("Clear Connections ID: %d succeed", connID)
		return
	}

	logger.Log().Debug("Clear Connections ID:  %d err", connID)
}
