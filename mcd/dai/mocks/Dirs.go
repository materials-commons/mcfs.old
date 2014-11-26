package mocks

import "github.com/stretchr/testify/mock"

import "github.com/materials-commons/mcfs/base/schema"

type Dirs struct {
	mock.Mock
}

func (m *Dirs) ByID(id string) (*schema.Directory, error) {
	ret := m.Called(id)

	r0 := ret.Get(0).(*schema.Directory)
	r1 := ret.Error(1)

	return r0, r1
}
func (m *Dirs) ByPath(path string) (*schema.Directory, error) {
	ret := m.Called(path)

	r0 := ret.Get(0).(*schema.Directory)
	r1 := ret.Error(1)

	return r0, r1
}
func (m *Dirs) Update(_a0 *schema.Directory) error {
	ret := m.Called(_a0)

	r0 := ret.Error(0)

	return r0
}
func (m *Dirs) Insert(_a0 *schema.Directory) (*schema.Directory, error) {
	ret := m.Called(_a0)

	r0 := ret.Get(0).(*schema.Directory)
	r1 := ret.Error(1)

	return r0, r1
}
func (m *Dirs) AddFiles(dir *schema.Directory, fileIDs ...string) error {
	ret := m.Called(dir, fileIDs)

	r0 := ret.Error(0)

	return r0
}
func (m *Dirs) RemoveFiles(dir *schema.Directory, fileIDs ...string) error {
	ret := m.Called(dir, fileIDs)

	r0 := ret.Error(0)

	return r0
}
