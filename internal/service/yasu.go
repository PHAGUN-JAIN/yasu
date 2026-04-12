package service

type YasuService struct {
}

func (s *YasuService) GetMessage() string {
	return "Hello from YasuService! I am GetYasu."
}