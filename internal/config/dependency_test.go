package config

import "testing"

func TestGetDependencyServer(t *testing.T) {
	config := Config{
		Service: "myapp",
		Image:   "myapp",
		Servers: ServersConfig{servers: map[string][]Server{
			"web": {
				{Host: "192.168.1.10", Labels: []string{"primary"}},
			},
			"db": {
				{Host: "192.168.1.20", Labels: []string{"database"}},
			},
		}},
	}

	tests := []struct {
		name     string
		dep      Dependency
		wantHost string
		wantRole string
		wantErr  bool
	}{
		{
			name:     "by host",
			dep:      Dependency{Image: "postgres:16", Host: "192.168.1.20"},
			wantHost: "192.168.1.20",
			wantRole: "db",
			wantErr:  false,
		},
		{
			name:     "by role",
			dep:      Dependency{Image: "postgres:16", Role: "db"},
			wantHost: "192.168.1.20",
			wantRole: "db",
			wantErr:  false,
		},
		{
			name:     "by labels",
			dep:      Dependency{Image: "postgres:16", Labels: []string{"database"}},
			wantHost: "192.168.1.20",
			wantRole: "db",
			wantErr:  false,
		},
		{
			name:     "default to primary",
			dep:      Dependency{Image: "postgres:16"},
			wantHost: "192.168.1.10",
			wantRole: "web",
			wantErr:  false,
		},
		{
			name:    "invalid host",
			dep:     Dependency{Image: "postgres:16", Host: "192.168.1.99"},
			wantErr: true,
		},
		{
			name:    "invalid role",
			dep:     Dependency{Image: "postgres:16", Role: "nonexistent"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, role, err := config.GetDependencyServer(tt.dep)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDependencyServer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if server.Host != tt.wantHost {
					t.Errorf("GetDependencyServer() host = %v, want %v", server.Host, tt.wantHost)
				}
				if role != tt.wantRole {
					t.Errorf("GetDependencyServer() role = %v, want %v", role, tt.wantRole)
				}
			}
		})
	}
}

