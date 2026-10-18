package standalone_storage

import (
	"github.com/Connor1996/badger"
	"github.com/pingcap-incubator/tinykv/kv/config"
	"github.com/pingcap-incubator/tinykv/kv/storage"
	"github.com/pingcap-incubator/tinykv/kv/util/engine_util"
	"github.com/pingcap-incubator/tinykv/proto/pkg/kvrpcpb"
)

// StandAloneStorage is an implementation of `Storage` for a single-node TinyKV instance. It does not
// communicate with other nodes and all data is stored locally.
type StandAloneStorage struct {
	// Your Data Here (1).
	engine *engine_util.Engines
	// txn    *badger.Txn
}

func NewStandAloneStorage(conf *config.Config) *StandAloneStorage {
	// Your Code Here (1).
	kv := engine_util.CreateDB(conf.DBPath, false)
	// defer kv.Close()
	return &StandAloneStorage{
		engine: engine_util.NewEngines(kv, nil, conf.DBPath, ""),
		// txn:    nil,
	}
}

func (s *StandAloneStorage) Start() error {
	// Your Code Here (1).
	return nil
}

func (s *StandAloneStorage) Stop() error {
	// Your Code Here (1).
	return s.engine.Close()
}

// Reader 应返回一个 StorageReader ，它支持在快照上进行 key/value 的点查和扫描操作。
// hint: 你应该使用 badger.Txn 来实现 Reader 函数，因为 badger 提供的事务处理器能够提供键和值的一致性快照。
// 也就是放到一个事务中, StorageReader 中的接口都有
func (s *StandAloneStorage) Reader(ctx *kvrpcpb.Context) (storage.StorageReader, error) {
	// Your Code Here (1).
	return NewStandAloneStorageReader(s), nil
}

type StandAloneStorageReader struct {
	txn *badger.Txn
}

func NewStandAloneStorageReader(s *StandAloneStorage) *StandAloneStorageReader {
	t := s.engine.Kv.NewTransaction(false)

	return &StandAloneStorageReader{
		t,
	}
}

func (sr *StandAloneStorageReader) GetCF(cf string, key []byte) ([]byte, error) {
	val, err := engine_util.GetCFFromTxn(sr.txn, cf, key)
	if err == badger.ErrKeyNotFound {
		// 没找到
		return nil, nil
	}
	// 其他正常返回
	return val, err
}

func (sr *StandAloneStorageReader) IterCF(cf string) engine_util.DBIterator {
	return engine_util.NewCFIterator(cf, sr.txn)
}

// 读迭代器, 不需要调用
func (sr *StandAloneStorageReader) Close() {
	// 关闭事务
	if sr.txn != nil {
		sr.txn.Discard()
	}
}

// Write 应提供一种方式，将一系列修改应用到内部状态，而在此情况下，内部状态就是一个 badger 实例。
// 一系列修改, 说明有多个, 那么应该使用事务.
func (s *StandAloneStorage) Write(ctx *kvrpcpb.Context, batch []storage.Modify) error {
	// Your Code Here (1).

	// t := s.engine.Kv.NewTransaction(true)
	// defer t.Discard()
	// defer t.Commit()

	var wb engine_util.WriteBatch
	for _, m := range batch {
		cf := m.Cf()
		key := m.Key()
		val := m.Value()

		// 查看代码提供了哪些 api 调用: 是调用 util.go依次处理, 还是使用 write_batch.go ?
		// 先使用 util.go 处理, 有 DeleteCF.  要判断 CF 类型.

		// 还是用 write_batch 吧
		wb.SetCF(cf, key, val)
	}

	return wb.WriteToDB(s.engine.Kv)
}
