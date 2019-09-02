package local

import (
	"github.com/infinivision/common/miscellaneous"
	"github.com/infinivision/store"
)

var rooms = []byte("tenant")

func New(db store.Store) *tenant {
	return &tenant{db}
}

func (t *tenant) AddRoom(name, number string) error {
	tx := t.db.NewTransaction()
	defer tx.Cancel()
	if err := tx.Set([]byte(name), []byte(number)); err != nil {
		return err
	}
	if err := Inc(tx, []byte(number)); err != nil {
		return err
	}
	if err := tx.Mset(rooms, []byte(name), []byte{}); err != nil {
		return err
	}
	return tx.Commit()
}

func (t *tenant) DelRoom(name string) ([]string, error) {
	var rs []string

	tx := t.db.NewTransaction()
	defer tx.Cancel()
	num, err := tx.Get([]byte(name))
	if err != nil {
		return nil, err
	}
	ks, _, err := tx.Mkvs([]byte(name))
	if err != nil {
		return nil, err
	}
	if err := tx.Del([]byte(name)); err != nil {
		return nil, err
	}
	if err := tx.Mdel(rooms, []byte(name)); err != nil {
		return nil, err
	}
	if err := tx.Mclear([]byte(name)); err != nil {
		return nil, err
	}
	for _, v := range ks {
		rs = append(rs, string(v))
	}
	n, err := Dec(tx, []byte(num))
	switch {
	case err != nil:
		return nil, err
	case n == 0:
		if ks, _, err = tx.Mkvs([]byte(num)); err != nil {
			return nil, err
		}
		if err := tx.Mclear([]byte(num)); err != nil {
			return nil, err
		}
		for _, v := range ks {
			if err := tx.Mdel(v, []byte(num)); err != nil {
				return nil, err
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return rs, nil
}

func (t *tenant) Rent(name, renter string) (string, error) {
	tx := t.db.NewTransaction()
	defer tx.Cancel()
	number, err := tx.Get([]byte(name))
	if err != nil {
		return "", err
	}
	if err := tx.Mset([]byte(name), []byte(renter), []byte{}); err != nil {
		return "", err
	}
	if err := tx.Mset([]byte(renter), []byte(name), []byte{}); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return string(number), nil
}

func (t *tenant) Recycle(name, renter string) error {
	tx := t.db.NewTransaction()
	defer tx.Cancel()
	if err := tx.Mdel([]byte(name), []byte(renter)); err != nil {
		return err
	}
	if err := tx.Mdel([]byte(renter), []byte(name)); err != nil {
		return err
	}
	return tx.Commit()
}

func (t *tenant) Rooms() ([]string, error) {
	var rs []string

	ks, _, err := t.db.Mkvs(rooms)
	if err != nil {
		return nil, err
	}
	for _, v := range ks {
		rs = append(rs, string(v))
	}
	return rs, nil
}

func (t *tenant) Renters(name string) ([]string, error) {
	var rs []string

	ks, _, err := t.db.Mkvs([]byte(name))
	if err != nil {
		return nil, err
	}
	for _, v := range ks {
		rs = append(rs, string(v))
	}
	return rs, nil
}

func (t *tenant) RenterRooms(renter string) ([]string, error) {
	var rs []string

	ks, _, err := t.db.Mkvs([]byte(renter))
	if err != nil {
		return nil, err
	}
	for _, v := range ks {
		rs = append(rs, string(v))
	}
	return rs, nil
}

func (t *tenant) RoomNumber(name string) (string, error) {
	v, err := t.db.Get([]byte(name))
	if err != nil {
		return "", err
	}
	return string(v), err
}

func Inc(tx store.Transaction, k []byte) error {
	v, err := tx.Get(k)
	switch {
	case err == nil:
		n, _ := miscellaneous.D32func(v)
		if err = tx.Set(k, miscellaneous.E32func(n+1)); err != nil {
			return err
		}
	case err == store.NotExist:
		if err = tx.Set(k, miscellaneous.E32func(1)); err != nil {
			return err
		}
	default:
		return err
	}
	return nil
}

func Dec(tx store.Transaction, k []byte) (uint32, error) {
	v, err := tx.Get(k)
	if err != nil {
		return 0, err
	}
	n, _ := miscellaneous.D32func(v)
	switch {
	case n == 1:
		if err = tx.Del(k); err != nil {
			return 0, err
		}
		return 0, nil
	default:
		if err = tx.Set(k, miscellaneous.E32func(n-1)); err != nil {
			return 0, err
		}
		return n - 1, nil
	}
}
