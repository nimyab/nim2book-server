package delete

type Service struct {
}

var service *Service

func New() *Service {
	service = &Service{}
	return service
}
