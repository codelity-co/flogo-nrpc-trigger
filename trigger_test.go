package nrpc

import (
	"github.com/project-flogo/core/data/metadata"
	"encoding/json"
	"testing"

	"github.com/project-flogo/core/action"
	"github.com/project-flogo/core/support"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/support/test"
	"github.com/project-flogo/core/trigger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/nats-io/gnatsd/server"
	natsserver "github.com/nats-io/nats-server/test"
)

const testConfig string = `{
	"id": "flogo-timer",
	"ref": "github.com/project-flogo/contrib/trigger/kafka",
	"settings": {
	  "brokerUrls": "localhost:9092"
	},
	"handlers": [
	  {
			"action":{
				"id":"dummy"
			},
			"settings": {
		  	"topic": "syslog"
			}
	  }
	]
	
}`

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

	ref := support.GetRef(&Trigger{})
	f := trigger.GetFactory(ref)
	assert.NotNil(suite.T(), f, "Should return factory instance")

}

func (suite *TriggerTestSuite) TestFactoryMetadata() {

	ref := support.GetRef(&Trigger{})
	f := trigger.GetFactory(ref)
	assert.NotNil(suite.T(), f, "Should return factory instance")

	m := f.Metadata()
	assert.NotNil(suite.T(), m, "Factory should return its metadata")
	
}

func (suite *TriggerTestSuite) TestTriggerInitialize() {

	s := RunServerOnPort(4222)
	defer s.Shutdown()

	ref := support.GetRef(&Trigger{})
	f := trigger.GetFactory(ref)
	assert.NotNil(suite.T(), f, "Should return factory instance")

	// Test trigger configuration
	config := &trigger.Config{}
	err := json.Unmarshal([]byte(suite.testConfig), config)
	assert.Nil(suite.T(), err, "Invalid trigger config")

	// Create dummy action
	actions := map[string]action.Action{"dummy": test.NewDummyAction(func() {
		//do nothing
	})}

	// Test trigger instance
	trg, err := test.InitTrigger(f, config, actions)
	assert.Nil(suite.T(), err, "InitTrigger return error")
	assert.NotNil(suite.T(), trg, "Should return trigger instance")

}


func (suite *TriggerTestSuite) TestGetNatsConnAuthOpts() {
	config := &trigger.Config{}
	err := json.Unmarshal([]byte(suite.testConfig), config)
	assert.Nil(suite.T(), err, "Invalid trigger config")

	s := &Settings{}
	err = metadata.MapToStruct(config.Settings, s, true)
	assert.Nil(suite.T(), err, "MapToStruct error when converting json to Settings")

	_, err = getNatsConnAuthOpts(s)
	assert.Nil(suite.T(), err, "getNatsConnAuthOpts error")
}

func (suite *TriggerTestSuite) TestGetNatsConnReconnectOpts() {
	
	config := &trigger.Config{}
	err := json.Unmarshal([]byte(suite.testConfig), config)
	assert.Nil(suite.T(), err, "Invalid trigger config")

	s := &Settings{}
	err = metadata.MapToStruct(config.Settings, s, true)
	assert.Nil(suite.T(), err, "MapToStruct error when converting json to Settings")

	_, err = getNatsConnAuthOpts(s)
	assert.Nil(suite.T(), err, "getNatsConnAuthOpts error")
}

func (suite *TriggerTestSuite) TestGetNatsConnSslConfigOpts() {
	config := &trigger.Config{}
	err := json.Unmarshal([]byte(suite.testConfig), config)
	assert.Nil(suite.T(), err, "Invalid trigger config")

	s := &Settings{}
	err = metadata.MapToStruct(config.Settings, s, true)
	assert.Nil(suite.T(), err, "MapToStruct error when converting json to Settings")

	_, err = getNatsConnSslConfigOpts(s)
	assert.Nil(suite.T(), err, "getNatsConnSslConfigOpts error")
}

func (suite *TriggerTestSuite) TestHandlerGetConnection() {
	s := RunServerOnPort(4222)
	defer s.Shutdown()

	config := &trigger.Config{}
	err := json.Unmarshal([]byte(suite.testConfig), config)
	assert.Nil(suite.T(), err, "Invalid trigger config")
	
	triggerSettings := &Settings{}
	err = metadata.MapToStruct(config.Settings, triggerSettings, true)
	assert.Nil(suite.T(), err, "MapToStruct error when converting json to Settings")

	h := &Handler{
		logger: log.RootLogger(),
		triggerSettings: triggerSettings,
	}

	err = h.getConnection()
	assert.Nil(suite.T(), err, "getConnection error")
	assert.NotNil(suite.T(), h.natsConn, "Cannot get NATS connection")
}

func (suite *TriggerTestSuite) TestResolveObject() {
	_, err := resolveObject(map[string]interface{}{
		"test1": "=$property[TEST]",
	})
	assert.NotNil(suite.T(), err, "resolveObject error")
	assert.Equal(suite.T(), "failed to resolve Property: 'TEST', ensure that property is configured in the application", err.Error())

	_, err = resolveObject(map[string]interface{}{
		"test1": "=$env[TEST]",
	})
	assert.NotNil(suite.T(), err, "resolveObject error")
	assert.Equal(suite.T(), "failed to resolve Environment Variable: 'TEST', ensure that variable is configured", err.Error())
} 

func (suite *TriggerTestSuite) TestHandlerHandleMessage() {
	suite.T().Skip("Pending")
}

func TestTriggerTestSuite(t *testing.T) {
	suite.Run(t, new(TriggerTestSuite))
}

func RunServerWithOptions(opts *server.Options) *server.Server {
	return natsserver.RunServer(opts)
}

func RunServerOnPort(port int) *server.Server {
	opts := natsserver.DefaultTestOptions
	opts.Port = port
	return RunServerWithOptions(&opts)
}