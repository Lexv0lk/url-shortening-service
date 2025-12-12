package domain

//region UrlExistingError

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
