## Usage for custom-domain test suite

### Prepare a secret with cloud credentials to manage DNS.

Create the secret in the default namespace:

```
kubectl create secret generic google-credentials -n default --from-file=serviceaccount.json=serviceaccount.json
```

### Set the environment variables with custom domain

- `TEST_CUSTOM_DOMAIN` - set this environment variable with your desired custom domain.
- `TEST_DOMAIN` - set this environment variable with your installed by default Kyma domain.

After exporting these domains, run `make setup-custom-domain` to finish the default test setup.


### Run the tests

To start the test suite, run:

```
make test-custom-domain
```
