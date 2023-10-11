// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2021 Datadog, Inc.

package constants

const (
	// CIJobName indicates job name.
	CIJobName = "ci.job.name"

	// CIJobURL indicates job URL.
	CIJobURL = "ci.job.url"

	// CIPipelineID indicates pipeline ID.
	CIPipelineID = "ci.pipeline.id"

	// CIPipelineName indicates pipeline name.
	CIPipelineName = "ci.pipeline.name"

	// CIPipelineNumber indicates pipeline number.
	CIPipelineNumber = "ci.pipeline.number"

	// CIPipelineURL indicates pipeline URL.
	CIPipelineURL = "ci.pipeline.url"

	// CIProviderName indicates provider name.
	CIProviderName = "ci.provider.name"

	// CIStageName indicates stage name.
	CIStageName = "ci.stage.name"

	// CINodeName indicates the node name.
	CINodeName = "ci.node.name"

	// CINodeLabels indicates the node labels.
	CINodeLabels = "ci.node.labels"

	// CIWorkspacePath records an absolute path to the directory where the project has been checked out.
	CIWorkspacePath = "ci.workspace_path"

	// CIEnvVars contains env vars used to get the pipeline correlation ID
	CIEnvVars = "_dd.ci.env_vars"
)
