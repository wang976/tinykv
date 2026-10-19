package server

import (
	"context"

	"github.com/pingcap-incubator/tinykv/kv/storage"
	"github.com/pingcap-incubator/tinykv/proto/pkg/kvrpcpb"
)

// The functions below are Server's Raw API. (implements TinyKvServer).
// Some helper methods can be found in server.go in the current directory

// RawGet return the corresponding Get response based on RawGetRequest's CF and Key fields
// RawGet 根据 RawGetRequest 的 CF 和 Key 字段，返回对应的 Get 响应
func (server *Server) RawGet(_ context.Context, req *kvrpcpb.RawGetRequest) (*kvrpcpb.RawGetResponse, error) {
	// Your Code Here (1).
	reader, err := server.storage.Reader(req.Context)

	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// 没有表达 "key 不存在".  key 不存在时, value 应返回 nil.  需区分: "不存在" 和 "存在但 value 为空"
	val, err := reader.GetCF(req.Cf, req.Key)

	res := &kvrpcpb.RawGetResponse{}

	if val == nil && err == nil {
		// key 不存在, 在当前事务快照中不可见
		res.Value = nil
		res.NotFound = true

		return res, nil
	}

	if err == nil {
		// 正常返回, 不区分 value 是否为""
		res.Value = val
		res.NotFound = false

		return res, err
	}

	// val == nil, 其他 err: key 找到, 但读 value 失败, 真正的错误.
	return nil, err
}

// RawPut puts the target data into storage and returns the corresponding response
func (server *Server) RawPut(_ context.Context, req *kvrpcpb.RawPutRequest) (*kvrpcpb.RawPutResponse, error) {
	// Your Code Here (1).
	// Hint: Consider using Storage.Modify to store data to be modified
	// 先转换为 Modify 类型
	var p = storage.Put{
		Key:   req.Key,
		Value: req.Value,
		Cf:    req.Cf,
	}
	var mod []storage.Modify
	var m storage.Modify
	m.Data = p
	mod = append(mod, m)

	err := server.storage.Write(req.Context, mod)
	res := &kvrpcpb.RawPutResponse{}

	return res, err
}

// RawDelete delete the target data from storage and returns the corresponding response
func (server *Server) RawDelete(_ context.Context, req *kvrpcpb.RawDeleteRequest) (*kvrpcpb.RawDeleteResponse, error) {
	// Your Code Here (1).
	// Hint: Consider using Storage.Modify to store data to be deleted
	var d = storage.Delete{
		Key: req.Key,
		Cf:  req.Cf,
	}

	var mod []storage.Modify
	var m storage.Modify
	m.Data = d
	mod = append(mod, m)

	err := server.storage.Write(req.Context, mod)
	res := &kvrpcpb.RawDeleteResponse{}
	return res, err
}

// RawScan scan the data starting from the start key up to limit. and return the corresponding result
func (server *Server) RawScan(_ context.Context, req *kvrpcpb.RawScanRequest) (*kvrpcpb.RawScanResponse, error) {
	// Your Code Here (1).
	// Hint: Consider using reader.IterCF
	reader, err := server.storage.Reader(req.Context)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// 感觉不用 itercf 也可以啊, 还是使用吧
	iter := reader.IterCF(req.Cf)
	defer iter.Close()
	iter.Seek(req.StartKey)

	// 这里分配的是指针切片, 默认值是 nil
	// kv := make([]*kvrpcpb.KvPair, req.Limit)
	var kvs []*kvrpcpb.KvPair

	// 需要分配内存
	// for i := range kv {
	// 	kv[i] = &kvrpcpb.KvPair{}
	// }

	sum := 0
	for iter.Valid() {
		// limit 为 0
		if sum >= int(req.Limit) {
			break
		}

		item := iter.Item()

		var kv kvrpcpb.KvPair

		key := item.KeyCopy(nil)
		val, err := item.ValueCopy(nil)
		if err != nil {
			return nil, err
		}

		kv.Key = key
		kv.Value = val

		// 构造
		kvs = append(kvs, &kv)

		sum++
		iter.Next()
	}

	res := &kvrpcpb.RawScanResponse{
		Kvs: kvs,
	}

	return res, nil
}
