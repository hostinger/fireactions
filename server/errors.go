package server

type ErrEmptyField string

func (e ErrEmptyField) Error() string {
	return "Field is required: " + string(e)
}
