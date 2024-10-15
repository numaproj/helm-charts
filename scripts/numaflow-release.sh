# Download numaflow CRDs
wget -O charts/numaflow/crds/isbsvcs.yaml https://raw.githubusercontent.com/numaproj/numaflow/${NUMAFLOW_VERSION}/config/base/crds/full/numaflow.numaproj.io_interstepbufferservices.yaml
wget -O charts/numaflow/crds/pipelines.yaml https://raw.githubusercontent.com/numaproj/numaflow/${NUMAFLOW_VERSION}/config/base/crds/full/numaflow.numaproj.io_pipelines.yaml
wget -O charts/numaflow/crds/vertices.yaml https://raw.githubusercontent.com/numaproj/numaflow/${NUMAFLOW_VERSION}/config/base/crds/full/numaflow.numaproj.io_vertices.yaml
wget -O charts/numaflow/crds/monovertices.yaml https://raw.githubusercontent.com/numaproj/numaflow/${NUMAFLOW_VERSION}/config/base/crds/full/numaflow.numaproj.io_monovertices.yaml