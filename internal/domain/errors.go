package domain

//region UrlExistingError

// UrlExistingError is returned when attempting to create a mapping for a URL that already exists.
type UrlExistingError struct {
	Msg string
}

func (e *UrlExistingError) Error() string {
	return e.Msg
}

func (e *UrlExistingError) Is(target error) bool {
	_, ok := target.(*UrlExistingError)
	return ok
}

//endregion

//region UrlNonExistingError

// UrlNonExistingError is returned when a requested URL mapping does not exist in storage.
type UrlNonExistingError struct {
	Msg string
}

func (e *UrlNonExistingError) Error() string {
	return e.Msg
}

func (e *UrlNonExistingError) Is(target error) bool {
	_, ok := target.(*UrlNonExistingError)
	return ok
}

//endregion

//region InvalidUrlError

// InvalidUrlError is returned when the provided URL has invalid format or unsupported scheme.
type InvalidUrlError struct {
	Msg string
}

func (e *InvalidUrlError) Error() string {
	return e.Msg
}

func (e *InvalidUrlError) Is(target error) bool {
	_, ok := target.(*InvalidUrlError)
	return ok
}

//endregion

//region TokenNonExistingError

// TokenNonExistingError is returned when a requested URL token does not exist in storage.
type TokenNonExistingError struct {
	Msg string
}

func (e *TokenNonExistingError) Error() string {
	return e.Msg
}

func (e *TokenNonExistingError) Is(target error) bool {
	_, ok := target.(*TokenNonExistingError)
	return ok
}

//endregion
