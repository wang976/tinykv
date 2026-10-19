package engine_util

import (
	"os"

	"github.com/Connor1996/badger"
	"github.com/pingcap-incubator/tinykv/log"
)

// 核心作用: 把 TinyKV 使用的底层 Badger 数据库组织起来, 提供创建、批量写入、关闭和销毁的入口.
// 真正的数据库由 Badger 完成, 批量写入的具体逻辑在 WriteBatch 中.

// Engines keeps references to and data for the engines used by unistore.
// All engines are badger key/value databases.
// the Path fields are the filesystem path to where the data is stored.
type Engines struct {
	// Data, including data which is committed (i.e., committed across other nodes) and un-committed (i.e., only present
	// locally).
	Kv     *badger.DB // KV 数据库实例的指针
	KvPath string     // KV 数据库的磁盘目录
	// Metadata used by Raft.
	Raft     *badger.DB // 存储 Raft 日志、状态等信息的数据库实例指针
	RaftPath string     // Raft 数据库的磁盘目录
}

// KV 数据库和 Raft 数据库承担不同职责.前者保存服务使用的数据, 后者保存实现分布式一致性所需信息.

// 组装对象, 不打开数据库
func NewEngines(kvEngine, raftEngine *badger.DB, kvPath, raftPath string) *Engines {
	return &Engines{
		Kv:       kvEngine,
		KvPath:   kvPath,
		Raft:     raftEngine,
		RaftPath: raftPath,
	}
}

func (en *Engines) WriteKV(wb *WriteBatch) error {
	return wb.WriteToDB(en.Kv)
}

func (en *Engines) WriteRaft(wb *WriteBatch) error {
	return wb.WriteToDB(en.Raft)
}

// 关闭数据库, 保留磁盘数据.
func (en *Engines) Close() error {
	dbs := []*badger.DB{en.Kv, en.Raft}
	for _, db := range dbs {
		if db == nil {
			continue
		}
		if err := db.Close(); err != nil {
			return err
		}
	}
	return nil
}

// 关闭后删除数据库目录
func (en *Engines) Destroy() error {
	if err := en.Close(); err != nil {
		return err
	}
	if err := os.RemoveAll(en.KvPath); err != nil {
		return err
	}
	if err := os.RemoveAll(en.RaftPath); err != nil {
		return err
	}
	return nil
}

// 配置并打开底层数据库
// CreateDB creates a new Badger DB on disk at path.
func CreateDB(path string, raft bool) *badger.DB {
	opts := badger.DefaultOptions
	if raft {
		// Do not need to write blob for raft engine because it will be deleted soon.
		opts.ValueThreshold = 0
	}
	opts.Dir = path
	opts.ValueDir = opts.Dir
	if err := os.MkdirAll(opts.Dir, os.ModePerm); err != nil {
		log.Fatal(err)
	}
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
	return db
}
