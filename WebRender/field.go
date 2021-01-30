package WebRender

type Field struct {
	Tag         string      `html:"tag=input"`
	Name        string      `html:"name"`
	ID          string      `html:"id"`
	Value       interface{} `html:"value"`
	Class       string      `html:"class=form-controll"`
	Placeholder interface{} `html:"placeholder"`
	Subs        []Field
}

type BootrapInput struct {
	class       string `html:"class=form-group"`
	Name        string
	ID          string
	Placeholder string
	Input       Field
}

func (boot BootrapInput) String() string {

	return
}
