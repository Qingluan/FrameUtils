package task

type DefaultObj struct {
	Input []string
	Res   string
	Err   error
	Id    string
}

func (o DefaultObj) ID() string {
	return o.Id
}
func (o DefaultObj) String() string {
	return o.Res
}

func (o DefaultObj) Args() []string {
	return o.Input
}
func (o DefaultObj) Error() error {
	return o.Err
}
