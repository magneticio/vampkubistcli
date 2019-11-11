package client

import (
	"github.com/magneticio/vampkubistcli/models"
	"github.com/stretchr/testify/mock"
)

type RestClientMock struct {
	mock.Mock
}

func (m *RestClientMock) Login(username string, password string) (refreshToken string, accessToken string, err error) {
	args := m.Called(username, password)
	return args.Get(0).(string), args.Get(1).(string), args.Error(2)
}

func (m *RestClientMock) RefreshTokens() (refreshToken string, accessToken string, err error) {
	args := m.Called()
	return args.Get(0).(string), args.Get(1).(string), args.Error(2)
}

func (m *RestClientMock) Create(resourceName string, name string, source string, sourceType string, values map[string]string) (bool, error) {
	args := m.Called(resourceName, name, source, sourceType, values)
	return args.Get(0).(bool), args.Error(2)
}

func (m *RestClientMock) Update(resourceName string, name string, source string, sourceType string, values map[string]string) (bool, error) {
	args := m.Called(resourceName, name, source, sourceType, values)
	return args.Get(0).(bool), args.Error(1)
}

func (m *RestClientMock) PushMetricValueInternal(name string, source string, sourceType string, values map[string]string) (bool, error) {
	args := m.Called(name, source, sourceType, values)
	return args.Get(0).(bool), args.Error(1)
}

func (m *RestClientMock) PushMetricValue(name string, metricValue *models.MetricValue, values map[string]string) (bool, error) {
	args := m.Called(name, metricValue, values)
	return args.Get(0).(bool), args.Error(1)
}

func (m *RestClientMock) Delete(resourceName string, name string, values map[string]string) (bool, error) {
	args := m.Called(resourceName, name, values)
	return args.Get(0).(bool), args.Error(1)
}

func (m *RestClientMock) UpdatePassword(userName string, password string, values map[string]string) error {
	args := m.Called(userName, password, values)
	return args.Error(0)
}

func (m *RestClientMock) GetSpec(resourceName string, name string, outputFormat string, values map[string]string) (string, error) {
	args := m.Called(resourceName, name, outputFormat, values)
	return args.Get(0).(string), args.Error(1)
}

func (m *RestClientMock) Get(resourceName string, name string, outputFormat string, values map[string]string) (string, error) {
	args := m.Called(resourceName, name, outputFormat, values)
	return args.Get(0).(string), args.Error(1)
}

func (m *RestClientMock) List(resourceName string, outputFormat string, values map[string]string, simple bool) (string, error) {
	args := m.Called(resourceName, outputFormat, values, simple)
	return args.Get(0).(string), args.Error(1)
}

func (m *RestClientMock) UpdateUserPermission(username string, permission string, values map[string]string) (bool, error) {
	args := m.Called(username, permission, values)
	return args.Get(0).(bool), args.Error(1)
}

func (m *RestClientMock) RemovePermissionFromUser(username string, values map[string]string) (bool, error) {
	args := m.Called(username, values)
	return args.Get(0).(bool), args.Error(1)
}

func (m *RestClientMock) AddRoleToUser(username string, rolename string, values map[string]string) (bool, error) {
	args := m.Called(username, rolename, values)
	return args.Get(0).(bool), args.Error(1)
}

func (m *RestClientMock) RemoveRoleFromUser(username string, rolename string, values map[string]string) (bool, error) {
	args := m.Called(username, rolename, values)
	return args.Get(0).(bool), args.Error(1)
}

func (m *RestClientMock) Ping() (bool, error) {
	args := m.Called()
	return args.Get(0).(bool), args.Error(1)
}

func (m *RestClientMock) ReadNotifications(notifications chan<- models.Notification) error {
	args := m.Called()
	return args.Error(0)
}

func (m *RestClientMock) SendExperimentMetric(experimentName string, metricName string, experimentMetric *models.ExperimentMetric, values map[string]string) error {
	args := m.Called(experimentName, metricName, experimentMetric, values)
	return args.Error(0)
}

func (m *RestClientMock) GetSubsetMap(values map[string]string) (*models.DestinationsSubsetsMap, error) {
	args := m.Called(values)
	return args.Get(0).(*models.DestinationsSubsetsMap), args.Error(1)
}
