package tenant

type Tenant interface {
	AddRoom(string, string) error
	DelRoom(string) ([]string, error)

	Recycle(string, string) error
	Rent(string, string) (string, error)

	RoomNumber(string) (string, error)
}
