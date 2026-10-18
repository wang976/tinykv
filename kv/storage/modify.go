package storage

// Modify is a single modification to TinyKV's underlying storage.
type Modify struct {
	Data interface{} // 约定放 Put 或 Delete, 于是一个 []Modify 就能混合表示多种修改
}

// 两种修改操作
type Put struct {
	Key   []byte
	Value []byte
	Cf    string
}

type Delete struct {
	Key []byte
	Cf  string
}

func (m *Modify) Key() []byte {
	// 类型选择
	switch m.Data.(type) {
	case Put:
		return m.Data.(Put).Key // 类型断言
	case Delete:
		return m.Data.(Delete).Key
	}
	return nil
}

func (m *Modify) Value() []byte {
	// 带检查的类型断言
	if putData, ok := m.Data.(Put); ok {
		return putData.Value
	}

	return nil
}

func (m *Modify) Cf() string {
	switch m.Data.(type) {
	case Put:
		return m.Data.(Put).Cf
	case Delete:
		return m.Data.(Delete).Cf
	}
	return ""
}
