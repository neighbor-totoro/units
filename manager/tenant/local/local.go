package local

import (
	"github.com/infinivision/store"
)

func New(db store.Store) *tenant {
	return &tenant{db}
}

// 新增一间房间，房间名 -> 房间号
func (t *tenant) AddRoom(name, number string) error {
	return t.db.Set([]byte(name), []byte(number))
}

// 删除一间房间并通知所有租客
func (t *tenant) DelRoom(name string) ([]string, error) {
	var rs []string

	tx := t.db.NewTransaction()
	defer tx.Cancel()
	if _, err := tx.Get([]byte(name)); err != nil {
		return nil, err
	}
	ks, _, err := tx.Mkvs([]byte(name))
	if err != nil {
		return nil, err
	}
	if err := tx.Del([]byte(name)); err != nil {
		return nil, err
	}
	if err := tx.Mclear([]byte(name)); err != nil {
		return nil, err
	}
	for i, j := 0, len(ks); i < j; i++ {
		if err := tx.Mdel(ks[i], []byte(name)); err != nil {
			return nil, err
		}
		rs = append(rs, string(ks[i]))
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return rs, nil
}

// 将房间出租给用户
func (t *tenant) Rent(name, roomer string) (string, error) {
	tx := t.db.NewTransaction()
	defer tx.Cancel()
	number, err := tx.Get([]byte(name))
	if err != nil {
		return "", err
	}
	if err := tx.Mset([]byte(name), []byte(roomer), []byte{}); err != nil {
		return "", err
	}
	if err := tx.Mset([]byte(roomer), []byte(name), []byte{}); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return string(number), nil
}

// 回收某个租客的使用权
func (t *tenant) Recycle(name, roomer string) error {
	tx := t.db.NewTransaction()
	defer tx.Cancel()
	if err := tx.Mdel([]byte(name), []byte(roomer)); err != nil {
		return err
	}
	if err := tx.Mdel([]byte(roomer), []byte(name)); err != nil {
		return err
	}
	return tx.Commit()
}

func (t *tenant) RoomNumber(name string) (string, error) {
	v, err := t.db.Get([]byte(name))
	if err != nil {
		return "", err
	}
	return string(v), err
}
