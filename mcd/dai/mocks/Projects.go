package mocks

import "github.com/stretchr/testify/mock"

import "github.com/materials-commons/mcfs/base/dir"
import "github.com/materials-commons/mcfs/base/schema"

type Projects struct {
	mock.Mock
}

func (m *Projects) ByID(id string) (*schema.Project, error) {
	ret := m.Called(id)

	r0 := ret.Get(0).(*schema.Project)
	r1 := ret.Error(1)

	return r0, r1
}
func (m *Projects) ByName(name string) (*schema.Project, error) {
	ret := m.Called(name)

	r0 := ret.Get(0).(*schema.Project)
	r1 := ret.Error(1)

	return r0, r1
}
func (m *Projects) ForUser(user string) ([]schema.Project, error) {
	ret := m.Called(user)

	r0 := ret.Get(0).([]schema.Project)
	r1 := ret.Error(1)

	return r0, r1
}
func (m *Projects) Files(id string) ([]dir.FileInfo, error) {
	ret := m.Called(id)

	r0 := ret.Get(0).([]dir.FileInfo)
	r1 := ret.Error(1)

	return r0, r1
}
func (m *Projects) Update(_a0 *schema.Project) error {
	ret := m.Called(_a0)

	r0 := ret.Error(0)

	return r0
}
func (m *Projects) Insert(_a0 *schema.Project) (*schema.Project, error) {
	ret := m.Called(_a0)

	r0 := ret.Get(0).(*schema.Project)
	r1 := ret.Error(1)

	return r0, r1
}
func (m *Projects) AddDirectories(project *schema.Project, directoryIDs ...string) error {
	ret := m.Called(project, directoryIDs)

	r0 := ret.Error(0)

	return r0
}
