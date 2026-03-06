package headers

type Headers map[string]string

func NewHeaders() Headers {

}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
}
