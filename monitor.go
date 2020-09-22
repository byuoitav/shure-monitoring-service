package shure

// ReceiverMonitor is the interface met by an implementation of a
// receiver monitor
type ReceiverMonitor interface {
	Monitor(Receiver) error
}
