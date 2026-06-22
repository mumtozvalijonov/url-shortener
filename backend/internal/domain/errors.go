package domain

type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	ErrInvalidURL       Error = "invalid url"
	ErrShortCodeTaken   Error = "short code taken"
	ErrShortURLNotFound Error = "short url not found"
)
