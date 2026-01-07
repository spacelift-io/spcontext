package internal

import "errors"

func UnwrapError(err error) error {
	var unwrapped error

	for {
		unwrapped = errors.Unwrap(err)
		if unwrapped == nil {
			return err
		}
		err = unwrapped
	}
}
