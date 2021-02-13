package web

type CanCastHTML interface {
	ToHTML(...string) string
}

/*
TODO : this function
*/
func WrapDataToTableWithPages(castHTML CanCastHTML) string {
	return ""
}
