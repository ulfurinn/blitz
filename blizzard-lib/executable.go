package blizzard

type Executable struct {
	Exe      string
	Basename string
	Obsolete bool
}

func (e *Executable) release() {
	//fmt.Printf("releasing executable %v\n", *e)
	//os.Rename(e.Exe, fmt.Sprintf("blitz/deploy-old/%s", e.Basename))
	e.Obsolete = true
}
