package batch_handler

type Result struct {
	ToAck   []Index
	ToRetry []Index
	ToDlq   []Index
}

type Index struct {
	Idx int
	Err error
}

func NewResult() *Result {
	return &Result{
		ToAck:   make([]Index, 0),
		ToRetry: make([]Index, 0),
		ToDlq:   make([]Index, 0),
	}
}

func (r *Result) AddAck(idx int) {
	r.ToAck = append(r.ToAck, Index{idx, nil})
}

func (r *Result) AddRetry(idx int, err error) {
	r.ToRetry = append(r.ToRetry, Index{idx, err})
}

func (r *Result) AddDlq(idx int, err error) {
	r.ToDlq = append(r.ToDlq, Index{idx, err})
}
