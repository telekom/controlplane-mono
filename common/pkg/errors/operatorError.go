package operatorErrors

type OperatorError interface {
	error
	Retriable() bool
}
