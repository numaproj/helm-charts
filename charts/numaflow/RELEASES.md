## How to contribute

- Step 1: Clone [helm-chart](https://github.com/numaproj/helm-charts) repository in local and checkout a new branch from `main` branch.
- Step 2: Make the required changes in `charts/numaflow` accordingly.
    - Note: Make sure to update the version in `chart/numaflow/Chart.yaml` according to [sem versioning](https://semver.org/)
- Step 3: Get the changes merged in `main` branch
- Step 4: Create a helm package using `helm package charts/numaflow` from `main` branch on latest changes, it should create a package like `numaflow-x.x.x.tgz`.
- Step 5: Checkout to branch `gh-pages` and move `numaflow-x.x.x.tgz` in dir `numaflow`.
- Step 6: Run command `helm repo index numaflow --merge index.yaml --url https://numaproj.io/helm-charts`, it will generate a new `numaflow/index.yaml`
- Step 7: Run command `mv -f numaflow/index.yaml` to update the existing `index.yaml`
- Step 8: Fix the path of numaflow chart url from `https://numaproj.io/helm-charts/numaflow-x.x.x.tgz` to `https://numaproj.io/helm-charts/numaflow/numaflow-x.x.x.tgz` in `index.yaml`
- Step 9: Commit and raise the PR to get the changes merged.