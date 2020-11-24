package nrpc

import (
	"encoding/json"
	"testing"

	"github.com/project-flogo/core/action"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/support/test"
	"github.com/project-flogo/core/trigger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/nats-io/gnatsd/server"
	natsserver "github.com/nats-io/nats-server/test"
)

type TriggerTestSuite struct {
	suite.Suite
	testConfig string
}

func (suite *TriggerTestSuite) SetupTest() {
	suite.testConfig = `{
		"id": "flogo-nrpc-trigger",
		"ref": "github.com/codelity-co/flogo-nrpc-trigger",
		"settings": {
			"natsClusterUrls": "nats://localhost:4222"
		},
		"handlers": [
			{
				"action":{
					"id":"dummy"
				},
				"settings": {
					"port": 50051,
					"protoName": "echo",
					"protoFile": "../../gohub-core-apis/proto/gohub/echo.proto",
					"enableTLS": false
				}
			}
		]
	}`
}

func (suite *TriggerTestSuite) TestFactoryNew() {
	t := suite.T()

	ref := support.GetRef(&Trigger{})
	f := trigger.GetFactory(ref)
	assert.NotNil(t, f, "Should return factory instance")

}

func (suite *TriggerTestSuite) TestFactoryMetadata() {
	t := suite.T()

	ref := support.GetRef(&Trigger{})
	f := trigger.GetFactory(ref)
	assert.NotNil(t, f, "Should return factory instance")

	m := f.Metadata()
	assert.NotNil(t, m, "Factory should return its metadata")
	
}

func (suite *TriggerTestSuite) TestTriggerInitialize() {
	t := suite.T()

	s := RunServerWithOptions()
	defer s.Shutdown()

	ref := support.GetRef(&Trigger{})
	f := trigger.GetFactory(ref)
	assert.NotNil(t, f, "Should return factory instance")

	// Test trigger configuration
	config := &trigger.Config{}
	err := json.Unmarshal([]byte(suite.testConfig), config)
	assert.Nil(t, err, "Invalid trigger config")

	// Create dummy action
	actions := map[string]action.Action{"dummy": test.NewDummyAction(func() {
		//do nothing
	})}

	// Test trigger instance
	trg, err := test.InitTrigger(f, config, actions)
	assert.Nil(t, err, "InitTrigger return error")
	assert.NotNil(t, trg, "Should return trigger instance")

}


func (suite *TriggerTestSuite) TestGetNatsConnAuthOpts() {
	t := suite.T()

	config := &trigger.Config{}
	err := json.Unmarshal([]byte(suite.testConfig), config)
	assert.Nil(t, err, "Invalid trigger config")

	s := &Settings{}
	err = metadata.MapToStruct(config.Settings, s, true)
	assert.Nil(t, err, "MapToStruct error when converting json to Settings")

	_, err = getNatsConnAuthOpts(s)
	assert.Nil(t, err, "getNatsConnAuthOpts error")
}

func (suite *TriggerTestSuite) TestGetNatsConnReconnectOpts() {
	t := suite.T()

	config := &trigger.Config{}
	err := json.Unmarshal([]byte(suite.testConfig), config)
	assert.Nil(suite.T(), err, "Invalid trigger config")

	s := &Settings{}
	err = metadata.MapToStruct(config.Settings, s, true)
	assert.Nil(t, err, "MapToStruct error when converting json to Settings")

	_, err = getNatsConnAuthOpts(s)
	assert.Nil(t, err, "getNatsConnAuthOpts error")
}

func (suite *TriggerTestSuite) TestGetNatsConnSslConfigOpts() {
	t := suite.T()

	config := &trigger.Config{}
	err := json.Unmarshal([]byte(suite.testConfig), config)
	assert.Nil(t, err, "Invalid trigger config")

	s := &Settings{}
	err = metadata.MapToStruct(config.Settings, s, true)
	assert.Nil(t, err, "MapToStruct error when converting json to Settings")

	_, err = getNatsConnSslConfigOpts(s)
	assert.Nil(t, err, "getNatsConnSslConfigOpts error")
}

func (suite *TriggerTestSuite) TestHandlerGetConnection() {
	t := suite.T()

	s := RunServerWithOptions()
	defer s.Shutdown()

	config := &trigger.Config{}
	err := json.Unmarshal([]byte(suite.testConfig), config)
	assert.Nil(t, err, "Invalid trigger config")
	
	triggerSettings := &Settings{}
	err = metadata.MapToStruct(config.Settings, triggerSettings, true)
	assert.Nil(t, err, "MapToStruct error when converting json to Settings")

	h := &Handler{
		logger: log.RootLogger(),
		triggerSettings: triggerSettings,
	}

	err = h.getConnection()
	assert.Nil(t, err, "getConnection error")
	assert.NotNil(t, h.natsConn, "Cannot get NATS connection")
}

func (suite *TriggerTestSuite) TestResolveObject() {
	t := suite.T()

	_, err := resolveObject(map[string]interface{}{
		"test1": "=$property[TEST]",
	})
	assert.NotNil(t, err, "resolveObject error")
	assert.Equal(t, "failed to resolve Property: 'TEST', ensure that property is configured in the application", err.Error())

	_, err = resolveObject(map[string]interface{}{
		"test1": "=$env[TEST]",
	})
	assert.NotNil(t, err, "resolveObject error")
	assert.Equal(t, "failed to resolve Environment Variable: 'TEST', ensure that variable is configured", err.Error())
} 

func (suite *TriggerTestSuite) TestHandlerHandleMessage() {
	suite.T().Skip("Pending")
}

func TestTriggerTestSuite(t *testing.T) {
	suite.Run(t, new(TriggerTestSuite))
}

func RunServerWithOptions() *server.Server {
	return natsserver.RunServer(&natsserver.DefaultTestOptions)
}