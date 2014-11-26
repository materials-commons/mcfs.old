package mocks

import "github.com/stretchr/testify/mock"

import "github.com/materials-commons/mcfs/base/schema"

type Groups struct {
	mock.Mock
}

func (m *Groups) ByID(id string) (*schema.Group, error) {
	ret := m.Called(id)

	r0 := ret.Get(0).(*schema.Group)
	r1 := ret.Error(1)

	return r0, r1
}
func (m *Groups) Insert(_a0 *schema.Group) (*schema.Group, error) {
	ret := m.Called(_a0)

	r0 := ret.Get(0).(*schema.Group)
	r1 := ret.Error(1)

	return r0, r1
}
func (m *Groups) Delete(id string) error {
	ret := m.Called(id)

	r0 := ret.Error(0)

	return r0
}
func (m *Groups) HasAccess(owner string) bool {
	ret := m.Called(owner)

	r0 := ret.Get(0).(bool)

	return r0
}
