package p2p

var (
	Update int = 1
	Create int = 2
	Delete int = 3
)

type Data struct {
	FileID  string
	Content []byte
}

type Action struct {
	SenderID string
	Data     Data
	Type     int
}

func handeler(msg Data) {

}

func update(msg Data) {

}

func delete(msg Data) {

}

func create(msg Data) {

}
