package cmd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "echo command",
			args:    []string{"echo", "hello"},
			wantErr: false,
		},
		{
			name:    "invalid command",
			args:    []string{"invalidcmd123"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execCmd.SetContext(context.Background())

			err := execCmd.RunE(execCmd, tt.args)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
