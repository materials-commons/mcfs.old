package mocks

import "github.com/stretchr/testify/mock"

import "github.com/materials-commons/mcfs/base/schema"

type Users struct {
	mock.Mock
}

func (m *Users) ByID(id string) (*schema.User, error) {
	ret := m.Called(id)

	r0 := ret.Get(0).(*schema.User)
	r1 := ret.Error(1)

	return r0, r1
}
func (m *Users) ByAPIKey(apikey string) (*schema.User, error) {
	ret := m.Called(apikey)

	r0 := ret.Get(0).(*schema.User)
	r1 := ret.Error(1)

	return r0, r1
}
func (m *Users) All() ([]schema.User, error) {
	ret := m.Called()

	r0 := ret.Get(0).([]schema.User)
	r1 := ret.Error(1)

	return r0, r1
}
