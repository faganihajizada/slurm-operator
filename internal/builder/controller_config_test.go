// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"strings"
	"testing"

	slinkyv1beta1 "github.com/SlinkyProject/slurm-operator/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestBuilder_BuildControllerConfig(t *testing.T) {
	type fields struct {
		client client.Client
	}
	type args struct {
		controller *slinkyv1beta1.Controller
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
				client: fake.NewClientBuilder().
					WithObjects(&corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name: "prolog",
						},
						Data: map[string]string{
							"00-exit.sh": strings.Join([]string{
								"#!/usr/bin/sh",
								"exit 0",
							}, "\n"),
						},
					}).
					WithObjects(&corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name: "epilog",
						},
						Data: map[string]string{
							"00-exit.sh": strings.Join([]string{
								"#!/usr/bin/sh",
								"exit 0",
							}, "\n"),
						},
					}).
					Build(),
			},
			args: args{
				controller: &slinkyv1beta1.Controller{
					ObjectMeta: metav1.ObjectMeta{
						Name: "slurm",
					},
					Spec: slinkyv1beta1.ControllerSpec{
						ExtraConf: strings.Join([]string{
							"MinJobAge=2",
						}, "\n"),
						PrologScriptRefs: []slinkyv1beta1.ObjectReference{
							{Name: "prolog"},
						},
						EpilogScriptRefs: []slinkyv1beta1.ObjectReference{
							{Name: "epilog"},
						},
					},
				},
			},
		},
		{
			name: "with accounting, nodesets, config",
			fields: fields{
				client: fake.NewClientBuilder().
					WithObjects(&slinkyv1beta1.Accounting{
						ObjectMeta: metav1.ObjectMeta{
							Name: "slurm",
						},
					}).
					WithObjects(&slinkyv1beta1.Controller{
						ObjectMeta: metav1.ObjectMeta{
							Name: "slurm",
						},
					}).
					WithObjects(&slinkyv1beta1.NodeSet{
						ObjectMeta: metav1.ObjectMeta{
							Name: "slurm-foo",
						},
						Spec: slinkyv1beta1.NodeSetSpec{
							ControllerRef: slinkyv1beta1.ObjectReference{
								Name: "slurm",
							},
							ExtraConf: strings.Join([]string{
								"features=bar",
							}, " "),
							Partition: slinkyv1beta1.NodeSetPartition{
								Enabled: true,
							},
							Template: slinkyv1beta1.PodTemplate{
								PodSpecWrapper: slinkyv1beta1.PodSpecWrapper{
									PodSpec: corev1.PodSpec{
										Hostname: "foo-",
									},
								},
							},
						},
					}).
					WithObjects(&corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name: "slurm-config",
						},
						Data: map[string]string{
							cgroupConfFile: `# Override cgroup.conf
							CgroupPlugin=autodetect
							IgnoreSystemd=yes
							ConstrainCores=yes
							ConstrainRAMSpace=yes
							ConstrainDevices=yes
							ConstrainSwapSpace=yes`,
							"foo.conf": "Foo=bar",
						},
					}).
					Build(),
			},
			args: args{
				controller: &slinkyv1beta1.Controller{
					ObjectMeta: metav1.ObjectMeta{
						Name: "slurm",
					},
					Spec: slinkyv1beta1.ControllerSpec{
						AccountingRef: slinkyv1beta1.ObjectReference{
							Name: "slurm",
						},
						ConfigFileRefs: []slinkyv1beta1.ObjectReference{
							{Name: "slurm-config"},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := New(tt.fields.client)
			got, err := b.BuildControllerConfig(tt.args.controller)
			if (err != nil) != tt.wantErr {
				t.Errorf("Builder.BuildControllerConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			switch {
			case err != nil:
				return

			case got.Data[slurmConfFile] == "" && got.BinaryData[slurmConfFile] == nil:
				t.Errorf("got.Data[%s] = %v", slurmConfFile, got.Data[slurmConfFile])
			}
		})
	}
}

func Test_isCgroupEnabled(t *testing.T) {
	type args struct {
		cgroupConf string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "enabled",
			args: args{
				cgroupConf: "CgroupPlugin=autodetect",
			},
			want: true,
		},
		{
			name: "enabled, lowercase+multiline+comment",
			args: args{
				cgroupConf: `# Multiline file
cgroupplugin=autodetect # this is a comment
ignoresystemd=yes`,
			},
			want: true,
		},
		{
			name: "disabled",
			args: args{
				cgroupConf: "CgroupPlugin=disabled",
			},
			want: false,
		},
		{
			name: "disabled, lowercase+multiline+comment",
			args: args{
				cgroupConf: `# Multiline file
cgroupplugin=disabled # this is a comment
ignoresystemd=yes`,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isCgroupEnabled(tt.args.cgroupConf); got != tt.want {
				t.Errorf("isCgroupEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}
