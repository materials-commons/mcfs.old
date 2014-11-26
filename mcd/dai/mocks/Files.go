package mocks

import "github.com/stretchr/testify/mock"

import "github.com/materials-commons/mcfs/base/schema"

type Files struct {
	mock.Mock
}

func (m *Files) ByID(id string) (*schema.File, error) {
	ret := m.Called(id)

	r0 := ret.Get(0).(*schema.File)
	r1 := ret.Error(1)

	return r0, r1
}
func (m *Files) ByPath(name string) (*schema.File, error) {
	ret := m.Called(name)

	r0 := ret.Get(0).(*schema.File)
	r1 := ret.Error(1)

	return r0, r1
}
func (m *Files) ByPathChecksum(name string) ([]schema.File, error) {
	ret := m.Called(name)

	r0 := ret.Get(0).([]schema.File)
	r1 := ret.Error(1)

	return r0, r1
}
func (m *Files) ByPathPartials(name string) ([]schema.File, error) {
	ret := m.Called(name)

	r0 := ret.Get(0).([]schema.File)
	r1 := ret.Error(1)

	return r0, r1
}
func (m *Files) ByChecksum(checksum string) (*schema.File, error) {
	ret := m.Called(checksum)

	r0 := ret.Get(0).(*schema.File)
	r1 := ret.Error(1)

	return r0, r1
}
func (m *Files) MatchOn(key string) ([]schema.File, error) {
	ret := m.Called(key)

	r0 := ret.Get(0).([]schema.File)
	r1 := ret.Error(1)

	return r0, r1
}
func (m *Files) Hide(_a0 *schema.File) error {
	ret := m.Called(_a0)

	r0 := ret.Error(0)

	return r0
}
func (m *Files) Update(_a0 *schema.File) error {
	ret := m.Called(_a0)

	r0 := ret.Error(0)

	return r0
}
func (m *Files) Insert(file *schema.File) (*schema.File, error) {
	ret := m.Called(file)

	r0 := ret.Get(0).(*schema.File)
	r1 := ret.Error(1)

	return r0, r1
}
func (m *Files) InsertEntry(file *schema.File) (*schema.File, error) {
	ret := m.Called(file)

	r0 := ret.Get(0).(*schema.File)
	r1 := ret.Error(1)

	return r0, r1
}
func (m *Files) Delete(id string) error {
	ret := m.Called(id)

	r0 := ret.Error(0)

	return r0
}
func (m *Files) AddDirectories(file *schema.File, dirIDs ...string) error {
	ret := m.Called(file, dirIDs)

	r0 := ret.Error(0)

	return r0
}
