package lazyflag

type strerror string

func (s strerror) Error() string {
	return string(s)
}

const (
	NotFound       strerror = "not found resource"
	ErrDuplicate   strerror = "duplicate error"
	TypeNotSupport strerror = "type not support"
)
