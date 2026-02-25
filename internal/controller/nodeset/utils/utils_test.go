// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"
	"errors"
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	slinkyv1beta1 "github.com/SlinkyProject/slurm-operator/api/v1beta1"
	"github.com/SlinkyProject/slurm-operator/internal/builder/labels"
)

func newNodeSet(name string) *slinkyv1beta1.NodeSet {
	petMounts := []corev1.VolumeMount{
		{Name: "datadir", MountPath: "/tmp/zookeeper"},
	}
	podMounts := []corev1.VolumeMount{
		{Name: "home", MountPath: "/home"},
	}
	return newNodeSetWithVolumes(name, petMounts, podMounts)
}

func newNodeSetWithVolumes(name string, petMounts []corev1.VolumeMount, podMounts []corev1.VolumeMount) *slinkyv1beta1.NodeSet {
	mounts := petMounts
	mounts = append(mounts, podMounts...)
	claims := []corev1.PersistentVolumeClaim{}
	for _, m := range petMounts {
		claims = append(claims, newPVC(m.Name))
	}

	vols := []corev1.Volume{}
	for _, m := range podMounts {
		vols = append(vols, corev1.Volume{
			Name: m.Name,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: fmt.Sprintf("/tmp/%v", m.Name),
				},
			},
		})
	}

	template := slinkyv1beta1.PodTemplate{
		Metadata: slinkyv1beta1.Metadata{
			Labels: map[string]string{"foo": "bar"},
		},
		PodSpecWrapper: slinkyv1beta1.PodSpecWrapper{
			PodSpec: corev1.PodSpec{
				Volumes: vols,
			},
		},
	}

	return &slinkyv1beta1.NodeSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       slinkyv1beta1.NodeSetKind,
			APIVersion: slinkyv1beta1.NodeSetAPIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: corev1.NamespaceDefault,
			UID:       types.UID("test"),
		},
		Spec: slinkyv1beta1.NodeSetSpec{
			Replicas:    ptr.To[int32](1),
			ScalingMode: slinkyv1beta1.ScalingModeStatefulset,
			Slurmd: slinkyv1beta1.ContainerWrapper{
				Container: corev1.Container{
					Image:        "nginx",
					VolumeMounts: mounts,
				},
			},
			Template:             template,
			VolumeClaimTemplates: claims,
			UpdateStrategy: slinkyv1beta1.NodeSetUpdateStrategy{
				Type: slinkyv1beta1.RollingUpdateNodeSetStrategyType,
			},
			PersistentVolumeClaimRetentionPolicy: &slinkyv1beta1.NodeSetPersistentVolumeClaimRetentionPolicy{
				WhenScaled:  slinkyv1beta1.RetainPersistentVolumeClaimRetentionPolicyType,
				WhenDeleted: slinkyv1beta1.RetainPersistentVolumeClaimRetentionPolicyType,
			},
			RevisionHistoryLimit: ptr.To[int32](2),
		},
	}
}

func newPVC(name string) corev1.PersistentVolumeClaim {
	return corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: corev1.NamespaceDefault,
			Name:      name,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: *resource.NewQuantity(1, resource.BinarySI),
				},
			},
		},
	}
}

func newNodeSetWithControllerRef(name, controllerName string, uid types.UID) *slinkyv1beta1.NodeSet {
	ns := newNodeSet(name)
	ns.UID = uid
	ns.Spec.ControllerRef = slinkyv1beta1.ObjectReference{
		Namespace: corev1.NamespaceDefault,
		Name:      controllerName,
	}
	return ns
}

func newSetOwnerReferencesScheme() *runtime.Scheme {
	sch := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(sch))
	utilruntime.Must(slinkyv1beta1.AddToScheme(sch))
	return sch
}

func TestIsPodFromNodeSet(t *testing.T) {
	controller := &slinkyv1beta1.Controller{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		},
	}
	type args struct {
		nodeset *slinkyv1beta1.NodeSet
		pod     *corev1.Pod
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "From NodeSet",
			args: args{
				nodeset: newNodeSet("foo"),
				pod:     NewNodeSetStatefulSetPod(fake.NewFakeClient(), newNodeSet("foo"), controller, 0, ""),
			},
			want: true,
		},
		{
			name: "Not From NodeSet",
			args: args{
				nodeset: newNodeSet("foo"),
				pod:     NewNodeSetStatefulSetPod(fake.NewFakeClient(), newNodeSet("bar"), controller, 1, ""),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPodFromNodeSet(tt.args.nodeset, tt.args.pod); got != tt.want {
				t.Errorf("IsPodFromNodeSet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetOrdinal(t *testing.T) {
	controller := &slinkyv1beta1.Controller{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		},
	}
	type args struct {
		pod *corev1.Pod
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "foo-0",
			args: args{
				pod: NewNodeSetStatefulSetPod(fake.NewFakeClient(), newNodeSet("foo"), controller, 0, ""),
			},
			want: 0,
		},
		{
			name: "bar-1",
			args: args{
				pod: NewNodeSetStatefulSetPod(fake.NewFakeClient(), newNodeSet("bar"), controller, 1, ""),
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetOrdinal(tt.args.pod); got != tt.want {
				t.Errorf("GetOrdinal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetParentNameAndOrdinal(t *testing.T) {
	controller := &slinkyv1beta1.Controller{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		},
	}
	type args struct {
		pod *corev1.Pod
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 int
	}{
		{
			name: "foo-0",
			args: args{
				pod: NewNodeSetStatefulSetPod(fake.NewFakeClient(), newNodeSet("foo"), controller, 0, ""),
			},
			want:  "foo",
			want1: 0,
		},
		{
			name: "bar-1",
			args: args{
				pod: NewNodeSetStatefulSetPod(fake.NewFakeClient(), newNodeSet("bar"), controller, 1, ""),
			},
			want:  "bar",
			want1: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := GetParentNameAndOrdinal(tt.args.pod)
			if got != tt.want {
				t.Errorf("GetParentNameAndOrdinal() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetParentNameAndOrdinal() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestOrdinalGetPodName(t *testing.T) {
	type args struct {
		nodeset *slinkyv1beta1.NodeSet
		ordinal int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "foo-0",
			args: args{
				nodeset: newNodeSet("foo"),
				ordinal: 0,
			},
			want: "foo-0",
		},
		{
			name: "bar-1",
			args: args{
				nodeset: newNodeSet("bar"),
				ordinal: 1,
			},
			want: "bar-1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetOrdinalPodName(tt.args.nodeset, tt.args.ordinal); got != tt.want {
				t.Errorf("GetOrdinalPodName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetNodeName(t *testing.T) {
	controller := &slinkyv1beta1.Controller{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		},
	}
	type args struct {
		pod *corev1.Pod
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "foo-0",
			args: args{
				pod: NewNodeSetStatefulSetPod(fake.NewFakeClient(), newNodeSet("foo"), controller, 0, ""),
			},
			want: "foo-0",
		},
		{
			name: "bar-1",
			args: args{
				pod: NewNodeSetStatefulSetPod(fake.NewFakeClient(), newNodeSet("bar"), controller, 1, ""),
			},
			want: "bar-1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetNodeName(tt.args.pod); got != tt.want {
				t.Errorf("GetNodeName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsIdentityMatch(t *testing.T) {
	controller := &slinkyv1beta1.Controller{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		},
	}
	type args struct {
		nodeset *slinkyv1beta1.NodeSet
		pod     *corev1.Pod
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Match",
			args: args{
				nodeset: newNodeSet("foo"),
				pod:     NewNodeSetStatefulSetPod(fake.NewFakeClient(), newNodeSet("foo"), controller, 0, ""),
			},
			want: true,
		},
		{
			name: "Not Match",
			args: args{
				nodeset: newNodeSet("foo"),
				pod:     NewNodeSetStatefulSetPod(fake.NewFakeClient(), newNodeSet("bar"), controller, 1, ""),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsIdentityMatch(tt.args.nodeset, tt.args.pod); got != tt.want {
				t.Errorf("IsIdentityMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsStorageMatch(t *testing.T) {
	controller := &slinkyv1beta1.Controller{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		},
	}
	type args struct {
		nodeset *slinkyv1beta1.NodeSet
		pod     *corev1.Pod
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Match",
			args: args{
				nodeset: newNodeSet("foo"),
				pod:     NewNodeSetStatefulSetPod(fake.NewFakeClient(), newNodeSet("foo"), controller, 0, ""),
			},
			want: true,
		},
		{
			name: "Not Match",
			args: args{
				nodeset: newNodeSet("foo"),
				pod:     NewNodeSetStatefulSetPod(fake.NewFakeClient(), newNodeSet("bar"), controller, 1, ""),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsStorageMatch(tt.args.nodeset, tt.args.pod); got != tt.want {
				t.Errorf("IsStorageMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPersistentVolumeClaims(t *testing.T) {
	controller := &slinkyv1beta1.Controller{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		},
	}
	type args struct {
		nodeset *slinkyv1beta1.NodeSet
		pod     *corev1.Pod
	}
	tests := []struct {
		name string
		args args
		want map[string]corev1.PersistentVolumeClaim
	}{
		{
			name: "Without Claims",
			args: func() args {
				nodeset := &slinkyv1beta1.NodeSet{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: corev1.NamespaceDefault,
						Name:      "foo",
						Labels: map[string]string{
							"foo": "bar",
						},
					},
					Spec: slinkyv1beta1.NodeSetSpec{
						ScalingMode: slinkyv1beta1.ScalingModeStatefulset,
					},
				}
				return args{
					nodeset: nodeset,
					pod:     NewNodeSetStatefulSetPod(fake.NewFakeClient(), nodeset, controller, 0, ""),
				}
			}(),
			want: map[string]corev1.PersistentVolumeClaim{},
		},
		{
			name: "With Claims",
			args: args{
				nodeset: newNodeSet("foo"),
				pod:     NewNodeSetStatefulSetPod(fake.NewFakeClient(), newNodeSet("foo"), controller, 0, ""),
			},
			want: map[string]corev1.PersistentVolumeClaim{
				"datadir": {
					ObjectMeta: metav1.ObjectMeta{
						Namespace: corev1.NamespaceDefault,
						Name:      "datadir-foo-0",
						Labels:    labels.NewBuilder().WithWorkerSelectorLabels(newNodeSet("foo")).Build(),
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						Resources: corev1.VolumeResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: *resource.NewQuantity(1, resource.BinarySI),
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetPersistentVolumeClaims(tt.args.nodeset, tt.args.pod); !apiequality.Semantic.DeepEqual(got, tt.want) {
				t.Errorf("GetPersistentVolumeClaims() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPersistentVolumeClaimNameOrdinal(t *testing.T) {
	type args struct {
		nodeset       *slinkyv1beta1.NodeSet
		claim         *corev1.PersistentVolumeClaim
		paddedOrdinal string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Ordinal Zero",
			args: args{
				nodeset: newNodeSet("foo"),
				claim: &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: corev1.NamespaceDefault,
						Name:      "test",
					},
				},
				paddedOrdinal: "0",
			},
			want: "test-foo-0",
		},
		{
			name: "Non-Zero Ordinal",
			args: args{
				nodeset: newNodeSet("foo"),
				claim: &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: corev1.NamespaceDefault,
						Name:      "test",
					},
				},
				paddedOrdinal: "1",
			},
			want: "test-foo-1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetPersistentVolumeClaimNameOrdinal(tt.args.nodeset, tt.args.claim, tt.args.paddedOrdinal); got != tt.want {
				t.Errorf("GetPersistentVolumeClaimNameOrdinal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetOwnerReferences(t *testing.T) {
	sch := newSetOwnerReferencesScheme()
	listErr := errors.New("list failed")

	tests := []struct {
		name        string
		client      client.Client
		object      metav1.Object
		clusterName string
		wantErr     bool
		wantRefs    int
	}{
		{
			name: "no NodeSets in cluster",
			client: fake.NewClientBuilder().
				WithScheme(sch).
				WithObjects().
				Build(),
			object:      &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod", Namespace: corev1.NamespaceDefault}},
			clusterName: "my-cluster",
			wantErr:     false,
			wantRefs:    0,
		},
		{
			name: "one NodeSet matching cluster name",
			client: fake.NewClientBuilder().
				WithScheme(sch).
				WithObjects(newNodeSetWithControllerRef("nodeset-a", "my-cluster", "uid-a")).
				Build(),
			object:      &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod", Namespace: corev1.NamespaceDefault}},
			clusterName: "my-cluster",
			wantErr:     false,
			wantRefs:    1,
		},
		{
			name: "multiple NodeSets matching cluster name",
			client: fake.NewClientBuilder().
				WithScheme(sch).
				WithObjects(
					newNodeSetWithControllerRef("nodeset-a", "my-cluster", "uid-a"),
					newNodeSetWithControllerRef("nodeset-b", "my-cluster", "uid-b"),
				).
				Build(),
			object:      &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod", Namespace: corev1.NamespaceDefault}},
			clusterName: "my-cluster",
			wantErr:     false,
			wantRefs:    2,
		},
		{
			name: "NodeSets with different controller refs, only matching ones added",
			client: fake.NewClientBuilder().
				WithScheme(sch).
				WithObjects(
					newNodeSetWithControllerRef("nodeset-a", "my-cluster", "uid-a"),
					newNodeSetWithControllerRef("nodeset-b", "other-cluster", "uid-b"),
				).
				Build(),
			object:      &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod", Namespace: corev1.NamespaceDefault}},
			clusterName: "my-cluster",
			wantErr:     false,
			wantRefs:    1,
		},
		{
			name: "no NodeSets match cluster name",
			client: fake.NewClientBuilder().
				WithScheme(sch).
				WithObjects(newNodeSetWithControllerRef("nodeset-a", "other-cluster", "uid-a")).
				Build(),
			object:      &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod", Namespace: corev1.NamespaceDefault}},
			clusterName: "my-cluster",
			wantErr:     false,
			wantRefs:    0,
		},
		{
			name: "List returns error",
			client: fake.NewClientBuilder().
				WithScheme(sch).
				WithObjects(newNodeSetWithControllerRef("nodeset-a", "my-cluster", "uid-a")).
				WithInterceptorFuncs(interceptor.Funcs{
					List: func(_ context.Context, _ client.WithWatch, _ client.ObjectList, _ ...client.ListOption) error {
						return listErr
					},
				}).
				Build(),
			object:      &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod", Namespace: corev1.NamespaceDefault}},
			clusterName: "my-cluster",
			wantErr:     true,
			wantRefs:    0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := SetOwnerReferences(tt.client, ctx, tt.object, tt.clusterName)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetOwnerReferences() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			refs := tt.object.GetOwnerReferences()
			if len(refs) != tt.wantRefs {
				t.Errorf("SetOwnerReferences() owner refs count = %v, want %v", len(refs), tt.wantRefs)
			}
			for _, ref := range refs {
				if ref.Controller != nil && *ref.Controller {
					t.Errorf("SetOwnerReferences() set controller=true; expected non-controller owner ref")
				}
				if ref.BlockOwnerDeletion == nil || !*ref.BlockOwnerDeletion {
					t.Errorf("SetOwnerReferences() expected BlockOwnerDeletion=true")
				}
			}
		})
	}
}
