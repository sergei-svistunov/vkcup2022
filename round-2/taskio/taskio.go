package taskio

type Reader interface {
	DataCh() <-chan int64
	Err() error
}

type Writer interface {
	WriteData(r Reader) error
}
