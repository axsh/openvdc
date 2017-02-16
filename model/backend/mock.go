package backend

type MockClusterBackend struct{}

func (m *MockClusterBackend) Connect(dest []string) error {
	return nil
}
func (m *MockClusterBackend) Close() error {
	return nil
}
func (m *MockClusterBackend) Register(key string, value []byte) error {
	return nil
}
func (m *MockClusterBackend) Find(key string) ([]byte, error) {
	return []byte{}, nil
}
