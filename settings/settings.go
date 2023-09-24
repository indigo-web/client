package settings

type Settings struct {
	Body Body
}

type (
	Body struct {
		MaxChunkSize int64
	}
)
