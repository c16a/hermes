package mqtt

import (
	"errors"
	"github.com/c16a/hermes/lib/auth"
	"github.com/c16a/hermes/lib/config"
	"github.com/c16a/hermes/lib/persistence"
	"github.com/eclipse/paho.golang/packets"
	"io"
	"io/ioutil"
	"reflect"
	"sync"
	"testing"
)

func TestNewServerContext(t *testing.T) {
	type args struct {
		config *config.Config
	}
	tests := []struct {
		name    string
		args    args
		want    *ServerContext
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewServerContext(tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewServerContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewServerContext() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServerContext_AddClient(t *testing.T) {
	type fields struct {
		connectedClientsMap map[string]*ConnectedClient
		mu                  *sync.RWMutex
		config              *config.Config
		authProvider        auth.AuthorisationProvider
		persistenceProvider persistence.Provider
	}
	type args struct {
		conn    io.Writer
		connect *packets.Connect
	}
	tests := []struct {
		name              string
		fields            fields
		args              args
		wantCode          byte
		wantSessionExists bool
		wantMaxQos        byte
	}{
		{
			"Auth failed",
			fields{
				make(map[string]*ConnectedClient, 0),
				&sync.RWMutex{},
				&config.Config{
					Server: &config.Server{
						MaxQos: 2,
					},
				},
				&MockAuthProvider{throwError: true},
				nil,
			},
			args{
				ioutil.Discard,
				&packets.Connect{
					CleanStart: true,
					ClientID:   "abcd",
				},
			},
			135,
			false,
			2,
		},
		{
			"Adding fresh client",
			fields{
				make(map[string]*ConnectedClient, 0),
				&sync.RWMutex{},
				&config.Config{
					Server: &config.Server{
						MaxQos: 2,
					},
				},
				nil,
				nil,
			},
			args{
				ioutil.Discard,
				&packets.Connect{
					CleanStart: true,
					ClientID:   "abcd",
				},
			},
			0,
			false,
			2,
		},
		{
			"Existing client asked to revive session (no persistence)",
			fields{
				map[string]*ConnectedClient{
					"abcd": {
						ClientID:   "abcd",
						Connection: ioutil.Discard,
					},
				},
				&sync.RWMutex{},
				&config.Config{
					Server: &config.Server{
						MaxQos: 2,
					},
				},
				nil,
				nil,
			},
			args{
				ioutil.Discard,
				&packets.Connect{
					CleanStart: false,
					ClientID:   "abcd",
				},
			},
			0,
			true,
			2,
		},
		{
			"Existing client asked to revive session (with persistence)",
			fields{
				map[string]*ConnectedClient{
					"abcd": {
						ClientID:   "abcd",
						Connection: ioutil.Discard,
					},
				},
				&sync.RWMutex{},
				&config.Config{
					Server: &config.Server{
						MaxQos: 2,
					},
				},
				nil,
				&MockPersistenceProvider{},
			},
			args{
				ioutil.Discard,
				&packets.Connect{
					CleanStart: false,
					ClientID:   "abcd",
				},
			},
			0,
			true,
			2,
		},
		{
			"Existing client asked for fresh session",
			fields{
				map[string]*ConnectedClient{
					"abcd": {
						ClientID:   "abcd",
						Connection: ioutil.Discard,
					},
				},
				&sync.RWMutex{},
				&config.Config{
					Server: &config.Server{
						MaxQos: 2,
					},
				},
				nil,
				nil,
			},
			args{
				ioutil.Discard,
				&packets.Connect{
					CleanStart: true,
					ClientID:   "abcd",
				},
			},
			0,
			false,
			2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ServerContext{
				connectedClientsMap: tt.fields.connectedClientsMap,
				mu:                  tt.fields.mu,
				config:              tt.fields.config,
				authProvider:        tt.fields.authProvider,
				persistenceProvider: tt.fields.persistenceProvider,
			}
			gotCode, gotSessionExists, gotMaxQos := ctx.AddClient(tt.args.conn, tt.args.connect)
			if gotCode != tt.wantCode {
				t.Errorf("AddClient() gotCode = %v, want %v", gotCode, tt.wantCode)
			}
			if gotSessionExists != tt.wantSessionExists {
				t.Errorf("AddClient() gotSessionExists = %v, want %v", gotSessionExists, tt.wantSessionExists)
			}
			if gotMaxQos != tt.wantMaxQos {
				t.Errorf("AddClient() gotSessionExists = %v, want %v", gotMaxQos, tt.wantMaxQos)
			}
		})
	}
}

func TestServerContext_Disconnect(t *testing.T) {
	type fields struct {
		connectedClientsMap map[string]*ConnectedClient
		mu                  *sync.RWMutex
		config              *config.Config
		authProvider        auth.AuthorisationProvider
		persistenceProvider persistence.Provider
	}
	type args struct {
		conn       io.Writer
		disconnect *packets.Disconnect
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			"Deleting clean client",
			fields{
				map[string]*ConnectedClient{
					"abcd": {
						ClientID:   "abcd",
						Connection: ioutil.Discard,
						IsClean:    true,
					},
				},
				&sync.RWMutex{},
				&config.Config{},
				nil,
				nil,
			},
			args{
				ioutil.Discard,
				&packets.Disconnect{},
			},
		},
		{
			"Deleting persisted client",
			fields{
				map[string]*ConnectedClient{
					"abcd": {
						ClientID:   "abcd",
						Connection: ioutil.Discard,
						IsClean:    false,
					},
				},
				&sync.RWMutex{},
				&config.Config{},
				nil,
				nil,
			},
			args{
				ioutil.Discard,
				&packets.Disconnect{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ServerContext{
				connectedClientsMap: tt.fields.connectedClientsMap,
				mu:                  tt.fields.mu,
				config:              tt.fields.config,
				authProvider:        tt.fields.authProvider,
				persistenceProvider: tt.fields.persistenceProvider,
			}
			ctx.Disconnect(tt.args.conn, tt.args.disconnect)
		})
	}
}

func TestServerContext_Publish(t *testing.T) {
	type fields struct {
		connectedClientsMap map[string]*ConnectedClient
		mu                  *sync.RWMutex
		config              *config.Config
		authProvider        auth.AuthorisationProvider
		persistenceProvider persistence.Provider
	}
	type args struct {
		publish *packets.Publish
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Publish to connected client",
			fields: fields{
				map[string]*ConnectedClient{
					"abcd": {
						ClientID:    "abcd",
						Connection:  ioutil.Discard,
						IsConnected: true,
						Subscriptions: map[string]packets.SubOptions{
							"foo": {},
						},
					},
				},
				&sync.RWMutex{},
				&config.Config{},
				nil,
				nil,
			},
			args: args{
				&packets.Publish{
					Topic:   "foo",
					Payload: []byte("Hello World"),
				},
			},
		},
		{
			name: "Publish to disconnected persistent client (no persistence)",
			fields: fields{
				map[string]*ConnectedClient{
					"abcd": {
						ClientID:    "abcd",
						Connection:  ioutil.Discard,
						IsConnected: false,
						IsClean:     false,
						Subscriptions: map[string]packets.SubOptions{
							"foo": {},
						},
					},
				},
				&sync.RWMutex{},
				&config.Config{},
				nil,
				nil,
			},
			args: args{
				&packets.Publish{
					Topic:   "foo",
					Payload: []byte("Hello World"),
				},
			},
		},
		{
			name: "Publish to disconnected persistent client (with persistence)",
			fields: fields{
				map[string]*ConnectedClient{
					"abcd": {
						ClientID:    "abcd",
						Connection:  ioutil.Discard,
						IsConnected: false,
						IsClean:     false,
						Subscriptions: map[string]packets.SubOptions{
							"foo": {},
						},
					},
				},
				&sync.RWMutex{},
				&config.Config{},
				nil,
				&MockPersistenceProvider{},
			},
			args: args{
				&packets.Publish{
					Topic:   "foo",
					Payload: []byte("Hello World"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ServerContext{
				connectedClientsMap: tt.fields.connectedClientsMap,
				mu:                  tt.fields.mu,
				config:              tt.fields.config,
				authProvider:        tt.fields.authProvider,
				persistenceProvider: tt.fields.persistenceProvider,
			}
			ctx.Publish(tt.args.publish)
		})
	}
}

func TestServerContext_Subscribe(t *testing.T) {
	type fields struct {
		connectedClientsMap map[string]*ConnectedClient
		mu                  *sync.RWMutex
		config              *config.Config
		authProvider        auth.AuthorisationProvider
		persistenceProvider persistence.Provider
	}
	type args struct {
		conn      io.Writer
		subscribe *packets.Subscribe
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []byte
	}{
		{
			"Subscribing QoS 0",
			fields{
				map[string]*ConnectedClient{
					"abcd": {
						ClientID:      "abcd",
						Connection:    ioutil.Discard,
						IsConnected:   false,
						IsClean:       false,
						Subscriptions: make(map[string]packets.SubOptions, 0),
					},
				},
				&sync.RWMutex{},
				&config.Config{
					Server: &config.Server{
						MaxQos: 2,
					},
				},
				nil,
				&MockPersistenceProvider{},
			},
			args{
				ioutil.Discard,
				&packets.Subscribe{
					Subscriptions: map[string]packets.SubOptions{
						"foo": {
							QoS: 0,
						},
					},
				},
			},
			[]byte{packets.SubackGrantedQoS0},
		},
		{
			"Subscribing QoS 1",
			fields{
				map[string]*ConnectedClient{
					"abcd": {
						ClientID:      "abcd",
						Connection:    ioutil.Discard,
						IsConnected:   false,
						IsClean:       false,
						Subscriptions: make(map[string]packets.SubOptions, 0),
					},
				},
				&sync.RWMutex{},
				&config.Config{
					Server: &config.Server{
						MaxQos: 2,
					},
				},
				nil,
				&MockPersistenceProvider{},
			},
			args{
				ioutil.Discard,
				&packets.Subscribe{
					Subscriptions: map[string]packets.SubOptions{
						"foo": {
							QoS: 1,
						},
					},
				},
			},
			[]byte{packets.SubackGrantedQoS1},
		},
		{
			"Subscribing QoS 2",
			fields{
				map[string]*ConnectedClient{
					"abcd": {
						ClientID:      "abcd",
						Connection:    ioutil.Discard,
						IsConnected:   false,
						IsClean:       false,
						Subscriptions: make(map[string]packets.SubOptions, 0),
					},
				},
				&sync.RWMutex{},
				&config.Config{
					Server: &config.Server{
						MaxQos: 2,
					},
				},
				nil,
				&MockPersistenceProvider{},
			},
			args{
				ioutil.Discard,
				&packets.Subscribe{
					Subscriptions: map[string]packets.SubOptions{
						"foo": {
							QoS: 2,
						},
					},
				},
			},
			[]byte{packets.SubackGrantedQoS2},
		},
		{
			"Subscribing to higher Qos",
			fields{
				map[string]*ConnectedClient{
					"abcd": {
						ClientID:      "abcd",
						Connection:    ioutil.Discard,
						IsConnected:   false,
						IsClean:       false,
						Subscriptions: make(map[string]packets.SubOptions, 0),
					},
				},
				&sync.RWMutex{},
				&config.Config{
					Server: &config.Server{
						MaxQos: 1,
					},
				},
				nil,
				&MockPersistenceProvider{},
			},
			args{
				ioutil.Discard,
				&packets.Subscribe{
					Subscriptions: map[string]packets.SubOptions{
						"foo": {
							QoS: 2,
						},
					},
				},
			},
			[]byte{packets.SubackImplementationspecificerror},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ServerContext{
				connectedClientsMap: tt.fields.connectedClientsMap,
				mu:                  tt.fields.mu,
				config:              tt.fields.config,
				authProvider:        tt.fields.authProvider,
				persistenceProvider: tt.fields.persistenceProvider,
			}
			if got := ctx.Subscribe(tt.args.conn, tt.args.subscribe); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Subscribe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServerContext_Unsubscribe(t *testing.T) {
	type fields struct {
		connectedClientsMap map[string]*ConnectedClient
		mu                  *sync.RWMutex
		config              *config.Config
		authProvider        auth.AuthorisationProvider
		persistenceProvider persistence.Provider
	}
	type args struct {
		conn        io.Writer
		unsubscribe *packets.Unsubscribe
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []byte
	}{
		{
			"Unsubscribe known topic",
			fields{
				map[string]*ConnectedClient{
					"abcd": {
						ClientID:    "abcd",
						Connection:  ioutil.Discard,
						IsConnected: false,
						IsClean:     false,
						Subscriptions: map[string]packets.SubOptions{
							"foo": {
								QoS: 1,
							},
						},
					},
				},
				&sync.RWMutex{},
				&config.Config{
					Server: &config.Server{
						MaxQos: 1,
					},
				},
				nil,
				&MockPersistenceProvider{},
			},
			args{
				conn: ioutil.Discard,
				unsubscribe: &packets.Unsubscribe{
					Topics:     []string{"foo", "bar"},
					Properties: nil,
					PacketID:   0,
				},
			},
			[]byte{packets.UnsubackSuccess, packets.UnsubackNoSubscriptionFound},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ServerContext{
				connectedClientsMap: tt.fields.connectedClientsMap,
				mu:                  tt.fields.mu,
				config:              tt.fields.config,
				authProvider:        tt.fields.authProvider,
				persistenceProvider: tt.fields.persistenceProvider,
			}
			if got := ctx.Unsubscribe(tt.args.conn, tt.args.unsubscribe); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Unsubscribe() = %v, want %v", got, tt.want)
			}
		})
	}
}

type MockPersistenceProvider struct {
}

func (m *MockPersistenceProvider) ReservePacketID(clientID string, packetID uint16) error {
	return nil
}

func (m *MockPersistenceProvider) FreePacketID(clientID string, packetID uint16) error {
	return nil
}

func (m *MockPersistenceProvider) CheckForPacketIdReuse(clientID string, packetID uint16) (bool, error) {
	return false, nil
}

func (m *MockPersistenceProvider) SaveForOfflineDelivery(clientId string, publish *packets.Publish) error {
	return nil
}

func (m *MockPersistenceProvider) GetMissedMessages(clientId string) ([]*packets.Publish, error) {
	return []*packets.Publish{
		{
			Topic:   "foo",
			Payload: []byte("Hello World"),
		},
	}, nil
}

type MockAuthProvider struct {
	throwError bool
}

func (m *MockAuthProvider) Validate(username string, password string) error {
	if m.throwError {
		return errors.New("some random error")
	}
	return nil
}
