package lib

import (
	"github.com/c16a/hermes/lib/auth"
	"github.com/c16a/hermes/lib/config"
	"github.com/c16a/hermes/lib/persistence"
	"github.com/eclipse/paho.golang/packets"
	"io"
	"io/ioutil"
	"log"
	"net"
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
	}{
		{
			"Adding fresh client",
			fields{
				make(map[string]*ConnectedClient, 0),
				&sync.RWMutex{},
				&config.Config{},
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
				&config.Config{},
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
				&config.Config{},
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
				&config.Config{},
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
			gotCode, gotSessionExists := ctx.AddClient(tt.args.conn, tt.args.connect)
			if gotCode != tt.wantCode {
				t.Errorf("AddClient() gotCode = %v, want %v", gotCode, tt.wantCode)
			}
			if gotSessionExists != tt.wantSessionExists {
				t.Errorf("AddClient() gotSessionExists = %v, want %v", gotSessionExists, tt.wantSessionExists)
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
		conn       net.Conn
		disconnect *packets.Disconnect
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
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
			log.Println(ctx)
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
		// TODO: Add test cases.
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
			log.Println(ctx)
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
		conn      net.Conn
		subscribe *packets.Subscribe
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []byte
	}{
		// TODO: Add test cases.
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
		conn        net.Conn
		unsubscribe *packets.Unsubscribe
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []byte
	}{
		// TODO: Add test cases.
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
