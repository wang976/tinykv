package engine_util

import (
	"github.com/Connor1996/badger"
	"github.com/golang/protobuf/proto"
	"github.com/pingcap/errors"
)

type WriteBatch struct {
	entries       []*badger.Entry
	size          int
	safePoint     int
	safePointSize int
	safePointUndo int
}

const (
	CfDefault string = "default"
	CfWrite   string = "write"
	CfLock    string = "lock"
)

var CFs [3]string = [3]string{CfDefault, CfWrite, CfLock}

func (wb *WriteBatch) Len() int {
	return len(wb.entries)
}

func (wb *WriteBatch) SetCF(cf string, key, val []byte) {
	wb.entries = append(wb.entries, &badger.Entry{
		Key:   KeyWithCF(cf, key),
		Value: val,
	})
	wb.size += len(key) + len(val)
}

func (wb *WriteBatch) DeleteMeta(key []byte) {
	wb.entries = append(wb.entries, &badger.Entry{
		Key: key,
	})
	wb.size += len(key)
}

func (wb *WriteBatch) DeleteCF(cf string, key []byte) {
	wb.entries = append(wb.entries, &badger.Entry{
		Key: KeyWithCF(cf, key),
	})
	wb.size += len(key)
}

func (wb *WriteBatch) SetMeta(key []byte, msg proto.Message) error {
	val, err := proto.Marshal(msg)
	if err != nil {
		return errors.WithStack(err)
	}
	wb.entries = append(wb.entries, &badger.Entry{
		Key:   key,
		Value: val,
	})
	wb.size += len(key) + len(val)
	return nil
}

func (wb *WriteBatch) SetSafePoint() {
	wb.safePoint = len(wb.entries)
	wb.safePointSize = wb.size
}

func (wb *WriteBatch) RollbackToSafePoint() {
	wb.entries = wb.entries[:wb.safePoint]
	wb.size = wb.safePointSize
}

// 把 WriteBatch 里攒下的所有修改，作为一次 badger 事务原子地写入数据库.
func (wb *WriteBatch) WriteToDB(db *badger.DB) error {
	if len(wb.entries) > 0 {
		err := db.Update(func(txn *badger.Txn) error {
			for _, entry := range wb.entries {
				var err1 error
				if len(entry.Value) == 0 {
					// 写入一个删除标记.  问题: 空值 -> 删除标记. 但 val 本意也可能为空.
					err1 = txn.Delete(entry.Key)
				} else {
					// 写入一个键值对
					err1 = txn.SetEntry(entry)
				}
				if err1 != nil {
					return err1
				}
			}
			return nil
		})
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (wb *WriteBatch) MustWriteToDB(db *badger.DB) {
	err := wb.WriteToDB(db)
	if err != nil {
		panic(err)
	}
}

func (wb *WriteBatch) Reset() {
	wb.entries = wb.entries[:0]
	wb.size = 0
	wb.safePoint = 0
	wb.safePointSize = 0
	wb.safePointUndo = 0
}
