// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"context"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	slinkyv1beta1 "github.com/SlinkyProject/slurm-operator/api/v1beta1"
	"github.com/SlinkyProject/slurm-operator/internal/builder/labels"
	"github.com/SlinkyProject/slurm-operator/internal/utils/config"
	"github.com/SlinkyProject/slurm-operator/internal/utils/structutils"
)

const (
	slurmConfFile  = "slurm.conf"
	cgroupConfFile = "cgroup.conf"
)

func (b *Builder) BuildControllerConfig(controller *slinkyv1beta1.Controller) (*corev1.ConfigMap, error) {
	ctx := context.TODO()

	accounting, err := b.refResolver.GetAccounting(ctx, controller.Spec.AccountingRef)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, err
		}
	}

	nodesetList, err := b.refResolver.GetNodeSetsForController(ctx, controller)
	if err != nil {
		return nil, err
	}

	configFilesList := &corev1.ConfigMapList{
		Items: make([]corev1.ConfigMap, 0, len(controller.Spec.ConfigFileRefs)),
	}
	for _, ref := range controller.Spec.ConfigFileRefs {
		cm := &corev1.ConfigMap{}
		key := types.NamespacedName{
			Namespace: controller.Namespace,
			Name:      ref.Name,
		}
		if err := b.client.Get(ctx, key, cm); err != nil {
			return nil, err
		}
		configFilesList.Items = append(configFilesList.Items, *cm)
	}
	cgroupEnabled := true
	hasCgroupConfFile := false
	for _, configMap := range configFilesList.Items {
		if contents, ok := configMap.Data[cgroupConfFile]; ok {
			hasCgroupConfFile = true
			cgroupEnabled = isCgroupEnabled(contents)
		}
	}

	prologScripts := []string{}
	for _, ref := range controller.Spec.PrologScriptRefs {
		cm := &corev1.ConfigMap{}
		key := types.NamespacedName{
			Namespace: controller.Namespace,
			Name:      ref.Name,
		}
		if err := b.client.Get(ctx, key, cm); err != nil {
			return nil, err
		}
		filenames := structutils.Keys(cm.Data)
		sort.Strings(filenames)
		prologScripts = filenames
	}

	epilogScripts := []string{}
	for _, ref := range controller.Spec.EpilogScriptRefs {
		cm := &corev1.ConfigMap{}
		key := types.NamespacedName{
			Namespace: controller.Namespace,
			Name:      ref.Name,
		}
		if err := b.client.Get(ctx, key, cm); err != nil {
			return nil, err
		}
		filenames := structutils.Keys(cm.Data)
		sort.Strings(filenames)
		epilogScripts = filenames
	}

	prologSlurmctldScripts := []string{}
	for _, ref := range controller.Spec.PrologSlurmctldScriptRefs {
		cm := &corev1.ConfigMap{}
		key := types.NamespacedName{
			Namespace: controller.Namespace,
			Name:      ref.Name,
		}
		if err := b.client.Get(ctx, key, cm); err != nil {
			return nil, err
		}
		filenames := structutils.Keys(cm.Data)
		sort.Strings(filenames)
		prologSlurmctldScripts = filenames
	}

	epilogSlurmctldScripts := []string{}
	for _, ref := range controller.Spec.EpilogSlurmctldScriptRefs {
		cm := &corev1.ConfigMap{}
		key := types.NamespacedName{
			Namespace: controller.Namespace,
			Name:      ref.Name,
		}
		if err := b.client.Get(ctx, key, cm); err != nil {
			return nil, err
		}
		filenames := structutils.Keys(cm.Data)
		sort.Strings(filenames)
		epilogSlurmctldScripts = filenames
	}

	opts := ConfigMapOpts{
		Key:      controller.ConfigKey(),
		Metadata: controller.Spec.Template.PodMetadata,
		Data: map[string]string{
			slurmConfFile: buildSlurmConf(controller, accounting, nodesetList, prologScripts, epilogScripts, prologSlurmctldScripts, epilogSlurmctldScripts, cgroupEnabled),
		},
	}
	if !hasCgroupConfFile {
		opts.Data[cgroupConfFile] = buildCgroupConf()
	}

	opts.Metadata.Labels = structutils.MergeMaps(opts.Metadata.Labels, labels.NewBuilder().WithControllerLabels(controller).Build())

	return b.BuildConfigMap(opts, controller)
}

// https://slurm.schedmd.com/slurm.conf.html
func buildSlurmConf(
	controller *slinkyv1beta1.Controller,
	accounting *slinkyv1beta1.Accounting,
	nodesetList *slinkyv1beta1.NodeSetList,
	prologScripts, epilogScripts []string,
	prologSlurmctldScripts, epilogSlurmctldScripts []string,
	cgroupEnabled bool,
) string {
	controllerHost := fmt.Sprintf("%s(%s)", controller.PrimaryName(), controller.ServiceFQDNShort())

	conf := config.NewBuilder()

	conf.AddProperty(config.NewPropertyRaw("#"))
	conf.AddProperty(config.NewPropertyRaw("### GENERAL ###"))
	conf.AddProperty(config.NewProperty("ClusterName", controller.ClusterName()))
	conf.AddProperty(config.NewProperty("SlurmUser", slurmUser))
	conf.AddProperty(config.NewProperty("SlurmctldHost", controllerHost))
	conf.AddProperty(config.NewProperty("SlurmctldPort", SlurmctldPort))
	conf.AddProperty(config.NewProperty("StateSaveLocation", clusterSpoolDir(controller.ClusterName())))
	conf.AddProperty(config.NewProperty("SlurmdUser", slurmdUser))
	conf.AddProperty(config.NewProperty("SlurmdPort", SlurmdPort))
	conf.AddProperty(config.NewProperty("SlurmdSpoolDir", slurmdSpoolDir))
	conf.AddProperty(config.NewProperty("ReturnToService", 2))
	conf.AddProperty(config.NewProperty("MaxNodeCount", 1024))
	conf.AddProperty(config.NewProperty("GresTypes", "gpu"))

	conf.AddProperty(config.NewPropertyRaw("#"))
	conf.AddProperty(config.NewPropertyRaw("### LOGGING ###"))
	conf.AddProperty(config.NewProperty("SlurmctldLogFile", slurmctldLogFilePath))
	conf.AddProperty(config.NewProperty("SlurmSchedLogFile", slurmctldLogFilePath))
	conf.AddProperty(config.NewProperty("SlurmdLogFile", slurmdLogFilePath))
	conf.AddProperty(config.NewProperty("LogTimeFormat", logTimeFormat))

	conf.AddProperty(config.NewPropertyRaw("#"))
	conf.AddProperty(config.NewPropertyRaw("### PLUGINS & PARAMETERS ###"))
	conf.AddProperty(config.NewProperty("AuthType", authType))
	conf.AddProperty(config.NewProperty("CredType", credType))
	conf.AddProperty(config.NewProperty("AuthAltTypes", authAltTypes))
	conf.AddProperty(config.NewProperty("AuthAltParameters", authAltParameters))
	conf.AddProperty(config.NewProperty("AuthInfo", authInfo))
	conf.AddProperty(config.NewProperty("CommunicationParameters", "block_null_hash"))
	conf.AddProperty(config.NewProperty("SelectTypeParameters", "CR_Core_Memory"))
	if cgroupEnabled {
		conf.AddProperty(config.NewProperty("SlurmctldParameters", "enable_configless,enable_stepmgr"))
		conf.AddProperty(config.NewProperty("ProctrackType", "proctrack/cgroup"))
		conf.AddProperty(config.NewProperty("PrologFlags", "Contain"))
		conf.AddProperty(config.NewProperty("TaskPlugin", "task/cgroup,task/affinity"))
	} else {
		conf.AddProperty(config.NewProperty("SlurmctldParameters", "enable_configless"))
		conf.AddProperty(config.NewProperty("TaskPlugin", "task/affinity"))
	}

	conf.AddProperty(config.NewPropertyRaw("#"))
	conf.AddProperty(config.NewPropertyRaw("### ACCOUNTING ###"))
	if accounting != nil {
		conf.AddProperty(config.NewProperty("AccountingStorageType", "accounting_storage/slurmdbd"))
		conf.AddProperty(config.NewProperty("AccountingStorageHost", accounting.ServiceKey().Name))
		conf.AddProperty(config.NewProperty("AccountingStoragePort", SlurmdbdPort))
		conf.AddProperty(config.NewProperty("AccountingStorageTRES", "gres/gpu"))
		if cgroupEnabled {
			conf.AddProperty(config.NewProperty("JobAcctGatherType", "jobacct_gather/cgroup"))
		}
	} else {
		conf.AddProperty(config.NewProperty("AccountingStorageType", "accounting_storage/none"))
		conf.AddProperty(config.NewProperty("JobAcctGatherType", "jobacct_gather/none"))
	}

	if len(prologSlurmctldScripts) > 0 || len(epilogSlurmctldScripts) > 0 {
		conf.AddProperty(config.NewPropertyRaw("#"))
		conf.AddProperty(config.NewPropertyRaw("### SLURMCTLD PROLOG & EPILOG ###"))
	}
	for _, filename := range prologSlurmctldScripts {
		scriptPath := path.Join(slurmEtcDir, filename)
		conf.AddProperty(config.NewProperty("PrologSlurmctld", scriptPath))
	}
	for _, filename := range epilogSlurmctldScripts {
		scriptPath := path.Join(slurmEtcDir, filename)
		conf.AddProperty(config.NewProperty("EpilogSlurmctld", scriptPath))
	}

	if len(prologScripts) > 0 || len(epilogScripts) > 0 {
		conf.AddProperty(config.NewPropertyRaw("#"))
		conf.AddProperty(config.NewPropertyRaw("### PROLOG & EPILOG ###"))
	}
	for _, filename := range prologScripts {
		conf.AddProperty(config.NewProperty("Prolog", filename))
	}
	for _, filename := range epilogScripts {
		conf.AddProperty(config.NewProperty("Epilog", filename))
	}

	if len(nodesetList.Items) > 0 {
		conf.AddProperty(config.NewPropertyRaw("#"))
		conf.AddProperty(config.NewPropertyRaw("### COMPUTE & PARTITION ###"))
	}
	for _, nodeset := range nodesetList.Items {
		name := nodeset.Name
		template := nodeset.Spec.Template.PodSpecWrapper
		if template.Hostname != "" {
			name = strings.Trim(template.Hostname, "-")
		}
		nodesetLine := []string{
			fmt.Sprintf("NodeSet=%v", name),
			fmt.Sprintf("Feature=%v", name),
		}
		nodesetLineRendered := strings.Join(nodesetLine, " ")
		conf.AddProperty(config.NewPropertyRaw(nodesetLineRendered))
		partition := nodeset.Spec.Partition
		if !partition.Enabled {
			continue
		}
		partitionLine := []string{
			fmt.Sprintf("PartitionName=%v", name),
			fmt.Sprintf("Nodes=%v", name),
			partition.Config,
		}
		partitionLineRendered := strings.Join(partitionLine, " ")
		conf.AddProperty(config.NewPropertyRaw(partitionLineRendered))
	}

	extraConf := controller.Spec.ExtraConf
	conf.AddProperty(config.NewPropertyRaw("#"))
	conf.AddProperty(config.NewPropertyRaw("### EXTRA CONFIG ###"))
	conf.AddProperty(config.NewPropertyRaw(extraConf))

	return conf.Build()
}

// https://slurm.schedmd.com/cgroup.conf.html
func buildCgroupConf() string {
	conf := config.NewBuilder()

	conf.AddProperty(config.NewProperty("CgroupPlugin", "cgroup/v2"))
	conf.AddProperty(config.NewProperty("IgnoreSystemd", "yes"))

	return conf.Build()
}

func isCgroupEnabled(cgroupConf string) bool {
	r := regexp.MustCompile(`(?im)^CgroupPlugin=disabled`)
	found := r.FindStringSubmatch(cgroupConf)
	return len(found) == 0
}
