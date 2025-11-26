// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"testing"

	slinkyv1beta1 "github.com/SlinkyProject/slurm-operator/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestBuilder_BuildWorkerSshHostKeys(t *testing.T) {
	type fields struct {
		client client.Client
	}
	type args struct {
		nodeset *slinkyv1beta1.NodeSet
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "default",
			fields: fields{
				client: fake.NewFakeClient(),
			},
			args: args{
				nodeset: &slinkyv1beta1.NodeSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "slurm",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := New(tt.fields.client)
			got, err := b.BuildWorkerSshHostKeys(tt.args.nodeset)
			if (err != nil) != tt.wantErr {
				t.Errorf("Builder.BuildWorkerSshHostKeys() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			switch {
			case got.Data[sshHostEcdsaKeyFile] == nil && got.StringData[sshHostEcdsaKeyFile] == "":
				t.Errorf("got.Data[%s] = %v", sshHostEcdsaKeyFile, got.Data[sshHostEcdsaKeyFile])
			case got.Data[sshHostEcdsaPubKeyFile] == nil && got.StringData[sshHostEcdsaPubKeyFile] == "":
				t.Errorf("got.Data[%s] = %v", sshHostEcdsaPubKeyFile, got.Data[sshHostEcdsaPubKeyFile])

			case got.Data[sshHostEd25519KeyFile] == nil && got.StringData[sshHostEd25519KeyFile] == "":
				t.Errorf("got.Data[%s] = %v", sshHostEd25519KeyFile, got.Data[sshHostEd25519KeyFile])
			case got.Data[sshHostEd25519PubKeyFile] == nil && got.StringData[sshHostEd25519PubKeyFile] == "":
				t.Errorf("got.Data[%s] = %v", sshHostEd25519PubKeyFile, got.Data[sshHostEd25519PubKeyFile])

			case got.Data[sshHostRsaKeyFile] == nil && got.StringData[sshHostRsaKeyFile] == "":
				t.Errorf("got.Data[%s] = %v", sshHostRsaKeyFile, got.Data[sshHostRsaKeyFile])
			case got.Data[sshHostRsaPubKeyFile] == nil && got.StringData[sshHostRsaPubKeyFile] == "":
				t.Errorf("got.Data[%s] = %v", sshHostRsaPubKeyFile, got.Data[sshHostRsaPubKeyFile])
			}
		})
	}
}
