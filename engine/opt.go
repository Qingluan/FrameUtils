package engine

type Option map[string]string

const (
	OPT_LEFTJOIN      = 0x000
	OPT_RIGHTJOIN     = 0x001
	OPT_JOINT_NO_SAME = 0x010
	// OPT_JOINT_USE_FIRST = "first"
	// JOIN_OPT            = "direction"
	// USE                 = "USE_OPT"
)
