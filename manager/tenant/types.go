package tenant

type Tenant interface {
	AddRoom(string, string) error
	DelRoom(string) ([]string, error)

	Recycle(string, string) error
	Rent(string, string) (string, error)

	Rooms() ([]string, error)
	Renters(string) ([]string, error)
	RenterRooms(string) ([]string, error)

	RoomNumber(string) (string, error)
}
