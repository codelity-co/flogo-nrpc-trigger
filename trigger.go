package nrpc

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/mapper"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/data/property"
	"github.com/project-flogo/core/data/resolve"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/trigger"

	nats "github.com/nats-io/nats.go"
)

var triggerMd = trigger.NewMetadata(&Settings{}, &Output{})
var resolver = resolve.NewCompositeResolver(map[string]resolve.Resolver{
	".":        &resolve.ScopeResolver{},
	"env":      &resolve.EnvResolver{},
	"property": &property.Resolver{},
	"loop":     &resolve.LoopResolver{},
})

func init() {
	_ = trigger.Register(&Trigger{}, &Factory{})
}

type Factory struct {
}

type Trigger struct {
	settings *Settings
	id       string
	natsHandlers []*Handler
	logger log.Logger
}

func (*Factory) New(config *trigger.Config) (trigger.Trigger, error) {
	s := &Settings{}
	err := metadata.MapToStruct(config.Settings, s, true)
	if err != nil {
		return nil, err
	}

	return &Trigger{id: config.Id, settings: s}, nil
}

func (f *Factory) Metadata() *trigger.Metadata {
	return triggerMd
}

// Metadata implements trigger.Trigger.Metadata
func (t *Trigger) Metadata() *trigger.Metadata {
	return triggerMd
}

func (t *Trigger) Initialize(ctx trigger.InitContext) error {

	t.logger = ctx.Logger()

	t.logger.Debugf("Trigger Settings: %v", t.settings)

	// Init handlers
	for _, handler := range ctx.GetHandlers() {

		// Create Trigger Handler
		natsHandler := &Handler{
			triggerSettings: t.settings,
			logger:          t.logger,
			stopChannel: make(chan bool),
			triggerHandler:  handler,
		}

		// Append handler
		t.natsHandlers = append(t.natsHandlers, natsHandler)
		t.logger.Debugf("Registered trigger handler successfully")
	}

	return nil
}


// Start implements util.Managed.Start
func (t *Trigger) Start() error {
	var err error 

	for _, handler := range t.natsHandlers {

		err = handler.getConnection()
		if err != nil {
			return err
		}
		
		protoName := t.settings.ProtoName
		protoName = strings.Split(protoName, ".")[0]

		// Register each serviceName + protoName
		if len(ServiceRegistery.ServerServices) != 0 {
			for k, service := range ServiceRegistery.ServerServices {
				servRegFlag := false
				if strings.Compare(k, protoName+service.ServiceInfo().ServiceName) == 0 {
					t.logger.Infof("Registered Proto [%v] and Service [%v]", protoName, service.ServiceInfo().ServiceName)
					service.RunRegisterServerService(handler.natsConn, t, handler)
					servRegFlag = true
				}
				if !servRegFlag {
					t.logger.Errorf("Proto [%s] and Service [%s] not registered", protoName, service.ServiceInfo().ServiceName)
					return fmt.Errorf("Proto [%s] and Service [%s] not registered", protoName, service.ServiceInfo().ServiceName)
				}
			}

		} else {
			t.logger.Error("nRPC server services not registered")
			return errors.New("nRPC server services not registered")
		}
	}
	return nil
}

// Stop implements util.Managed.Stop
func (t *Trigger) Stop() error {
	for _, handler := range t.natsHandlers {
		handler.stopChannel <- true
		close(handler.natsMsgChannel)
		close(handler.stopChannel)
		_ = handler.natsConn.Drain()
		handler.natsConn.Close()
	}
	return nil
}

// Handler is a NATS subject handler
type Handler struct {
	triggerSettings  *Settings
	logger           log.Logger
	natsConn         *nats.Conn
	natsMsgChannel   chan interface{}
	natsSubscription *nats.Subscription
	stopChannel      chan bool
	triggerHandler   trigger.Handler
}

func (h *Handler) getConnection() error {
	var err error

	// Get NATS Connection
	h.logger.Infof("Getting NATS connection...")
	nc, err := getNatsConnection(h.logger, h.triggerSettings)
	if err != nil {
		return err
	}
	h.natsConn = nc
	h.logger.Infof("Got NATS connection")
	h.natsMsgChannel = make(chan interface{}) // Create NATS message channel
	return nil
}

func (h *Handler) HandleMessage() {

	for {

		select {

		case done := <-h.stopChannel: // Receive message from Stop Channel

			if done {
				return
			}

		case nrpcData := <-h.natsMsgChannel: // Receive NATS Msg from NATS message channel

			var (
				err error
				// results map[string]interface{}
			)

			out := &Output{
				Data: nrpcData.(map[string]interface{}),
			}

			result, err := h.triggerHandler.Handle(context.Background(), out)
			if err != nil {
				h.logger.Errorf("Trigger handler error: %v", err)
				continue
			}
			
			h.natsMsgChannel <- result
		}
	}
}

func getNatsConnection(logger log.Logger, settings *Settings) (*nats.Conn, error) {
	var (
		err           error
		authOpts      []nats.Option
		reconnectOpts []nats.Option
		sslConfigOpts []nats.Option
		urlString     string
	)

	// // Check ClusterUrls
	// if err := checkClusterUrls(settings); err != nil {
	// 	return nil, err
	// }

	urlString = settings.NatsClusterUrls

	authOpts, err = getNatsConnAuthOpts(settings)
	if err != nil {
		return nil, err
	}

	reconnectOpts, err = getNatsConnReconnectOpts(settings)
	if err != nil {
		return nil, err
	}

	sslConfigOpts, err = getNatsConnSslConfigOpts(settings)
	if err != nil {
		return nil, err
	}

	natsOptions := append(authOpts, reconnectOpts...)
	natsOptions = append(natsOptions, sslConfigOpts...)

	// Check ConnName
	if len(settings.NatsConnName) > 0 {
		natsOptions = append(natsOptions, nats.Name(settings.NatsConnName))
	}

	return nats.Connect(urlString, natsOptions...)

}

// getNatsConnAuthOps return slice of nats.Option specific for NATS authentication
func getNatsConnAuthOpts(settings *Settings) ([]nats.Option, error) {
	opts := make([]nats.Option, 0)
	// Check auth setting

	if settings.NatsUserName != "" { // Check if usename is defined
	  // check if password is defined
		if settings.NatsUserPassword == "" {
			return nil, fmt.Errorf("Missing password")
		} else {
			// Create UserInfo NATS option
			opts = append(opts, nats.UserInfo(settings.NatsUserName, settings.NatsUserPassword))
		}
	} else if settings.NatsToken != "" { // Check if token is defined
		opts = append(opts, nats.Token(settings.NatsToken))
	} else if settings.NatsNkeySeedfile != "" { // Check if nkey seed file is defined
		nkey, err := nats.NkeyOptionFromSeed(settings.NatsNkeySeedfile)
		if err != nil {
			return nil, err
		}
		opts = append(opts, nkey)
	} else if settings.NatsCredentialFile != "" { // Check if credential file is defined
		opts = append(opts, nats.UserCredentials(settings.NatsCredentialFile))
	}
	return opts, nil
}

func getNatsConnReconnectOpts(settings *Settings) ([]nats.Option, error) {
	opts := make([]nats.Option, 0)

	// Enable autoReconnect
	if !settings.AutoReconnect {
		opts = append(opts, nats.NoReconnect())
	}
	
	// Max reconnect attempts
	if settings.MaxReconnects > 0 {
		opts = append(opts, nats.MaxReconnects(settings.MaxReconnects))
	}

	// Don't randomize
	if settings.EnableRandomReconnection {
		opts = append(opts, nats.DontRandomize())
	}

	// Reconnect wait in seconds
	if settings.ReconnectWait > 0 {
		duration, err := time.ParseDuration(fmt.Sprintf("%vs", settings.ReconnectWait))
		if err != nil {
			return nil, err
		}
		opts = append(opts, nats.ReconnectWait(duration))
	}

	// Reconnect buffer size in bytes
	if settings.ReconnectBufferSize > 0 {
		opts = append(opts, nats.ReconnectBufSize(settings.ReconnectBufferSize))
	}
	return opts, nil
}

func getNatsConnSslConfigOpts(settings *Settings) ([]nats.Option, error) {
	opts := make([]nats.Option, 0)

	if settings.CertFile != "" && settings.KeyFile != "" {
		// Skip verify
		if settings.SkipVerify {
			opts = append(opts, nats.Secure(&tls.Config{
				InsecureSkipVerify: settings.SkipVerify,
			}))
		}
		// CA Root
		if settings.CaFile != "" {
			opts = append(opts, nats.RootCAs(settings.CaFile))
			// Cert file
			if settings.CertFile != "" {
				if settings.KeyFile != "" {
					opts = append(opts, nats.ClientCert(settings.CertFile, settings.KeyFile))
				} else {
					return nil, fmt.Errorf("Missing keyFile setting")
				}
			} else {
				return nil, fmt.Errorf("Missing certFile setting")
			}
		} else {
			return nil, fmt.Errorf("Missing caFile setting")
		}
	}
	return opts, nil
}

func resolveObject(object map[string]interface{}) (map[string]interface{}, error) {
	var err error

	mapperFactory := mapper.NewFactory(resolver)
	valuesMapper, err := mapperFactory.NewMapper(object)
	if err != nil {
		return nil, err
	}

	objectValues, err := valuesMapper.Apply(data.NewSimpleScope(map[string]interface{}{}, nil))
	if err != nil {
		return nil, err
	}

	return objectValues, nil
}
