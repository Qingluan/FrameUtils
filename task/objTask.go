package task

type TaskObj interface {
	Args() []string
	ID() string
	String() string
	Error() error
	ToGo() string
	Path() string
}
