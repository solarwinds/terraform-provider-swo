package typex

// Zero returns the zero value for the type U.
func Zero[U any]() U {
	var zero U
	return zero
}
