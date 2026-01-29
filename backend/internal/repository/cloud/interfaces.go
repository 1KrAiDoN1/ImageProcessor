package cloud

type CloudStorageInterface interface {
	Upload(objectKey string, data []byte) error
	Download(objectKey string) ([]byte, error)
	Delete(objectKey string) error
}
