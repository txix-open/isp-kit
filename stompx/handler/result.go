package handler

type Result struct {
	Ack     bool
	Requeue bool
	Err     error
}

func Ack() Result {
	return Result{Ack: true}
}

func Requeue(err error) Result {
	return Result{Requeue: true, Err: err}
}
