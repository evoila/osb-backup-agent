package mutex

type Mutex chan bool

func (s Mutex) Acquire() {
	<-s
}

func (s Mutex) Release() {
	e := false
	s <- e
}
