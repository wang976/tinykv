package storage

import (
	"github.com/pingcap-incubator/tinykv/kv/util/engine_util"
	"github.com/pingcap-incubator/tinykv/proto/pkg/kvrpcpb"
)

// Storage represents the internal-facing server part of TinyKV, it handles sending and receiving from other
// TinyKV nodes. As part of that responsibility, it also reads and writes data to disk (or semi-permanent memory).
type Storage interface {
	Start() error                                       // 启动存储组件
	Stop() error                                        // 停止存储组件
	Write(ctx *kvrpcpb.Context, batch []Modify) error   // 执行一批修改
	Reader(ctx *kvrpcpb.Context) (StorageReader, error) // 获取读取数据的对象.  -> 关联 StorageReader()
}

type StorageReader interface {
	// When the key doesn't exist, return nil for the value
	GetCF(cf string, key []byte) ([]byte, error)
	IterCF(cf string) engine_util.DBIterator
	Close()
}
